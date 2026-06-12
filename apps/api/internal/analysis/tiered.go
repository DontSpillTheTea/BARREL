package analysis

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/ai"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocr"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocr/providers"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/rules"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/validators"
)

type TieredInput struct {
	Filename       string
	ContentType    string
	ImageBytes     []byte
	ExpectedFields models.ExpectedLabelFields
	BeverageType   string

	OcrManager   *ocr.Manager
	TextParser   *ai.TextParserProvider
	AIProvider   ai.Provider
	Catalog      *rules.Catalog
	OcrProvider  string

	EscalationEnabled       bool
	OcrMinConfidence        float64
	FieldMinConfidence      float64
	GovWarningMinSimilarity float64
}

func AnalyzeTiered(ctx context.Context, input TieredInput) models.LabelAnalysisResult {
	startTime := time.Now()
	timings := &models.PipelineTimings{}
	var escalationReasons []string

	res := models.LabelAnalysisResult{
		Filename:          input.Filename,
		BeverageType:      input.BeverageType,
		RequestedProvider: "tiered",
		ExpectedFields:    input.ExpectedFields,
		OverallStatus:     models.StatusMatch,
		Warnings:          []string{},
	}

	// Stage 1: OCR
	ocrStart := time.Now()
	ocrProviderName := input.OcrProvider
	if ocrProviderName == "" || ocrProviderName == "tiered" {
		ocrProviderName = "azure_vision_ocr"
	}

	ocrInput := providers.ExtractInput{
		Filename:    input.Filename,
		ContentType: input.ContentType,
		Data:        input.ImageBytes,
	}
	ocrRes, ocrErr := input.OcrManager.Extract(ctx, ocrInput, ocrProviderName)
	timings.OcrTimeMs = time.Since(ocrStart).Milliseconds()

	if ocrErr != nil || ocrRes == nil || ocrRes.Status == "error" {
		log.Printf("Tiered: OCR failed, escalating to ai_native")
		escalationReasons = append(escalationReasons, "OCR extraction failed")
		return escalateToAINative(ctx, input, res, timings, escalationReasons, startTime)
	}

	res.OCR = ocrRes
	res.OCRText = ocrRes.Text
	log.Printf("Tiered: OCR completed in %dms, confidence=%.1f, text_length=%d", timings.OcrTimeMs, ocrRes.MeanConfidence, len(ocrRes.Text))

	// Check OCR confidence
	if ocrRes.MeanConfidence > 0 && ocrRes.MeanConfidence < input.OcrMinConfidence*100 {
		escalationReasons = append(escalationReasons, fmt.Sprintf("OCR confidence %.1f%% below threshold %.0f%%", ocrRes.MeanConfidence, input.OcrMinConfidence*100))
	}

	// Stage 2: Text parser (cheap LLM, text-only)
	parseStart := time.Now()
	var candidates *models.AINativeExtraction
	if input.TextParser != nil && ocrRes.Text != "" {
		var parseErr error
		candidates, parseErr = input.TextParser.ParseText(ctx, ocrRes.Text, input.BeverageType, input.ExpectedFields)
		timings.TextParseTimeMs = time.Since(parseStart).Milliseconds()
		if parseErr != nil {
			log.Printf("Tiered: text parser failed: %v, escalating", parseErr)
			escalationReasons = append(escalationReasons, "Text parser failed: "+parseErr.Error())
			return escalateToAINative(ctx, input, res, timings, escalationReasons, startTime)
		}
		log.Printf("Tiered: text parser completed in %dms", timings.TextParseTimeMs)
	} else if ocrRes.Text == "" {
		escalationReasons = append(escalationReasons, "OCR returned empty text")
		return escalateToAINative(ctx, input, res, timings, escalationReasons, startTime)
	} else {
		escalationReasons = append(escalationReasons, "Text parser not configured")
		return escalateToAINative(ctx, input, res, timings, escalationReasons, startTime)
	}

	// Stage 3: Build extracted fields and run deterministic validators
	validationStart := time.Now()
	res.ExtractedFields = models.ExtractedLabelFields{
		BrandName:              candidates.BrandName.Value,
		ClassType:              candidates.ClassType.Value,
		AlcoholContent:         candidates.AlcoholContent.ABV,
		NetContents:            candidates.NetContents.Value,
		GovernmentWarningFound: candidates.GovernmentWarning.Present,
		ProducerBottler:        candidates.ProducerOrBottler.Value,
		CountryOfOrigin:        candidates.CountryOfOrigin.Value,
	}

	res.AISecondRead = &models.AISecondRead{
		Eligible:   true,
		Used:       true,
		Provider:   "text_parser",
		Status:     "success",
		Candidates: *candidates,
	}
	res.AIEscalation = models.AIEscalation{
		Eligible: true,
		Used:     true,
		Provider: "text_parser",
		Reason:   "OCR text parsed by cheap LLM",
	}

	res = AnalyzeAI(res, input.Catalog)
	timings.ValidationTimeMs = time.Since(validationStart).Milliseconds()

	// Set source on all fields
	for i := range res.Fields {
		res.Fields[i].Source = "ocr_text"
	}

	// Stage 4: Confidence gate — check for escalation triggers
	escalationReasons = append(escalationReasons, checkEscalationTriggers(res, candidates, input)...)

	if len(escalationReasons) > 0 && input.EscalationEnabled && input.AIProvider != nil {
		log.Printf("Tiered: escalating to ai_native for %d reasons: %v", len(escalationReasons), escalationReasons)
		return escalateToAINative(ctx, input, res, timings, escalationReasons, startTime)
	}

	// Fast path: no escalation needed
	timings.TotalTimeMs = time.Since(startTime).Milliseconds()
	res.ProviderPath = "ocr_only"
	res.Timings = timings
	res.ProcessingTimeMs = timings.TotalTimeMs
	log.Printf("Tiered: fast path completed in %dms", timings.TotalTimeMs)
	return res
}

func checkEscalationTriggers(res models.LabelAnalysisResult, candidates *models.AINativeExtraction, input TieredInput) []string {
	var reasons []string

	optionalFields := map[string]bool{"Producer/Bottler": true, "Country of Origin": true}
	for _, f := range res.Fields {
		if optionalFields[f.Field] {
			continue
		}
		if f.Status == models.StatusMismatch {
			reasons = append(reasons, fmt.Sprintf("%s is Mismatch — AI vision will confirm", f.Field))
		}
		if f.Status == models.StatusMissingOnLabel && f.Expected != nil && f.Expected != "" {
			reasons = append(reasons, fmt.Sprintf("%s missing from OCR text", f.Field))
		}
		if f.Status == models.StatusUncertain {
			reasons = append(reasons, fmt.Sprintf("%s is Uncertain from OCR", f.Field))
		}
	}

	// Government warning specific checks
	if res.ExpectedFields.GovernmentWarningPresent {
		if !candidates.GovernmentWarning.Present {
			reasons = append(reasons, "Government warning not detected in OCR text")
		} else if candidates.GovernmentWarning.VerbatimText != "" {
			comparison := validators.CompareGovernmentWarningVerbatim(candidates.GovernmentWarning.VerbatimText)
			if comparison.Similarity < input.GovWarningMinSimilarity {
				reasons = append(reasons, fmt.Sprintf("Government warning similarity %.0f%% below threshold %.0f%%", comparison.Similarity*100, input.GovWarningMinSimilarity*100))
			}
		}
	}

	// Low field confidence from parser
	fields := []struct {
		name string
		conf float64
	}{
		{"brand_name", candidates.BrandName.Confidence},
		{"class_type", candidates.ClassType.Confidence},
		{"alcohol_content", candidates.AlcoholContent.Confidence},
		{"net_contents", candidates.NetContents.Confidence},
	}
	for _, f := range fields {
		if f.conf > 0 && f.conf < input.FieldMinConfidence {
			reasons = append(reasons, fmt.Sprintf("%s parser confidence %.0f%% below threshold %.0f%%", f.name, f.conf*100, input.FieldMinConfidence*100))
		}
	}

	return reasons
}

func escalateToAINative(ctx context.Context, input TieredInput, prelimRes models.LabelAnalysisResult, timings *models.PipelineTimings, reasons []string, startTime time.Time) models.LabelAnalysisResult {
	if input.AIProvider == nil {
		timings.TotalTimeMs = time.Since(startTime).Milliseconds()
		prelimRes.ProviderPath = "ocr_only"
		prelimRes.Escalated = true
		prelimRes.EscalationReasons = reasons
		prelimRes.Timings = timings
		prelimRes.ProcessingTimeMs = timings.TotalTimeMs
		prelimRes.Warnings = append(prelimRes.Warnings, "Escalation was triggered but AI-native provider is not configured.")
		return prelimRes
	}

	aiStart := time.Now()
	parserInput := ai.SecondReadInput{
		Filename:       input.Filename,
		ContentType:    http.DetectContentType(input.ImageBytes),
		ImageBytes:     input.ImageBytes,
		OCRText:        prelimRes.OCRText,
		ExpectedFields: input.ExpectedFields,
		BeverageType:   input.BeverageType,
		InitialResult:  prelimRes,
	}

	aiRes, err := input.AIProvider.SecondRead(ctx, parserInput)
	timings.AINativeTimeMs = time.Since(aiStart).Milliseconds()
	timings.TotalTimeMs = time.Since(startTime).Milliseconds()

	if err != nil || aiRes == nil {
		log.Printf("Tiered: ai_native escalation failed: %v", err)
		prelimRes.ProviderPath = "ocr_then_ai_native"
		prelimRes.Escalated = true
		prelimRes.EscalationReasons = reasons
		prelimRes.Timings = timings
		prelimRes.ProcessingTimeMs = timings.TotalTimeMs
		if err != nil {
			prelimRes.Warnings = append(prelimRes.Warnings, "AI-native escalation failed: "+err.Error())
		}
		return prelimRes
	}

	// Rebuild result from AI-native vision
	res := models.LabelAnalysisResult{
		Filename:          input.Filename,
		BeverageType:      input.BeverageType,
		RequestedProvider: "tiered",
		ExpectedFields:    input.ExpectedFields,
		OverallStatus:     models.StatusMatch,
		Warnings:          prelimRes.Warnings,
		OCRText:           prelimRes.OCRText,
		OCR:               prelimRes.OCR,
	}

	res.AISecondRead = aiRes
	res.AIEscalation = models.AIEscalation{
		Eligible: true,
		Used:     true,
		Provider: aiRes.Provider,
		Reason:   "Escalated from OCR tier",
	}
	res.ExtractedFields = models.ExtractedLabelFields{
		BrandName:              aiRes.Candidates.BrandName.Value,
		ClassType:              aiRes.Candidates.ClassType.Value,
		AlcoholContent:         aiRes.Candidates.AlcoholContent.ABV,
		NetContents:            aiRes.Candidates.NetContents.Value,
		GovernmentWarningFound: aiRes.Candidates.GovernmentWarning.Present,
		ProducerBottler:        aiRes.Candidates.ProducerOrBottler.Value,
		CountryOfOrigin:        aiRes.Candidates.CountryOfOrigin.Value,
	}

	if aiRes.Candidates.ImageQualityFlags != nil {
		res.ImageQualityFlags = aiRes.Candidates.ImageQualityFlags
	}

	res = AnalyzeAI(res, input.Catalog)

	for i := range res.Fields {
		res.Fields[i].Source = "ai_native"
	}

	res.ProviderPath = "ocr_then_ai_native"
	res.Escalated = true
	res.EscalationReasons = reasons
	res.Timings = timings
	res.ProcessingTimeMs = timings.TotalTimeMs

	log.Printf("Tiered: ai_native escalation completed in %dms (total %dms)", timings.AINativeTimeMs, timings.TotalTimeMs)
	return res
}
