package analysis

import (
	"fmt"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/rules"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/validators"
)

func AnalyzeAI(res models.LabelAnalysisResult, catalog *rules.Catalog) models.LabelAnalysisResult {
	if res.AIEscalation.Provider == "" {
		return res
	}

	prefix := ""
	if res.BeverageType == "distilled_spirits" {
		prefix = "spirits_"
	} else if res.BeverageType == "wine" {
		prefix = "wine_"
	} else if res.BeverageType == "malt_beverages" {
		prefix = "malt_"
	}

	var aiCandidates *models.AINativeExtraction
	if res.AISecondRead != nil {
		aiCandidates = &res.AISecondRead.Candidates
	}

	res.Fields = []models.FieldCheckResult{}

	// Pass image quality flags through
	if aiCandidates != nil && len(aiCandidates.ImageQualityFlags) > 0 {
		res.ImageQualityFlags = aiCandidates.ImageQualityFlags
	}

	// 1. Brand Name — tolerant fuzzy matching (PRD: case/punctuation differences are non-blocking)
	brandFound := res.ExtractedFields.BrandName
	brandAIConf := 0.0
	if aiCandidates != nil {
		brandAIConf = aiCandidates.BrandName.Confidence
	}
	brandField := compareFuzzyField("Brand Name", res.ExpectedFields.BrandName, brandFound, brandAIConf, 0.85, 0.70)
	brandField.Rule = getRule(catalog, prefix+"brand_name")
	res.Fields = append(res.Fields, brandField)

	// 2. Class/Type — tolerant fuzzy matching
	classFound := res.ExtractedFields.ClassType
	classAIConf := 0.0
	if aiCandidates != nil {
		classAIConf = aiCandidates.ClassType.Confidence
	}
	classField := compareFuzzyField("Class/Type", res.ExpectedFields.ClassType, classFound, classAIConf, 0.85, 0.70)
	classField.Rule = getRule(catalog, prefix+"class_type")
	res.Fields = append(res.Fields, classField)

	// 3. Alcohol Content — numeric tolerance per beverage type (27 CFR § 5.37, § 4.36, § 7.71)
	alcFound := res.ExtractedFields.AlcoholContent
	alcAIConf := 0.0
	if aiCandidates != nil {
		alcAIConf = aiCandidates.AlcoholContent.Confidence
	}
	alcStatus, alcSim, alcExpl := validators.CompareABV(res.ExpectedFields.AlcoholContent, alcFound, res.BeverageType)
	alcConf := int(alcSim * 100)
	if alcFound == "" || alcFound == "Not found" {
		alcConf = 40
	}
	if alcStatus == models.StatusMatch {
		alcConf = 90
	}
	if alcAIConf > 0 && alcAIConf < 0.70 && alcStatus == models.StatusMatch {
		alcStatus = models.StatusUncertain
		alcExpl = "Alcohol content text matches but AI extraction confidence is low."
	}
	res.Fields = append(res.Fields, models.FieldCheckResult{
		Field:        "Alcohol Content",
		Expected:     res.ExpectedFields.AlcoholContent,
		Found:        alcFound,
		Status:       alcStatus,
		Confidence:   alcConf,
		Similarity:   alcSim,
		AIConfidence: alcAIConf,
		Explanation:  alcExpl,
		Rule:         getRule(catalog, prefix+"alcohol_content"),
	})

	// 4. Net Contents — unit-normalized comparison
	netFound := res.ExtractedFields.NetContents
	netAIConf := 0.0
	if aiCandidates != nil {
		netAIConf = aiCandidates.NetContents.Confidence
	}
	netStatus, netSim, netExpl := validators.CompareNetContents(res.ExpectedFields.NetContents, netFound)
	netConf := int(netSim * 100)
	if netFound == "" || netFound == "Not found" {
		netConf = 40
	}
	if netStatus == models.StatusMatch {
		netConf = 90
	}
	if netAIConf > 0 && netAIConf < 0.70 && netStatus == models.StatusMatch {
		netStatus = models.StatusUncertain
		netExpl = "Net contents match but AI extraction confidence is low."
	}
	res.Fields = append(res.Fields, models.FieldCheckResult{
		Field:        "Net Contents",
		Expected:     res.ExpectedFields.NetContents,
		Found:        netFound,
		Status:       netStatus,
		Confidence:   netConf,
		Similarity:   netSim,
		AIConfidence: netAIConf,
		Explanation:  netExpl,
		Rule:         getRule(catalog, prefix+"net_contents"),
	})

	if valid, warning := validators.ValidateContainerSize(netFound, res.BeverageType); !valid && warning != "" {
		res.Warnings = append(res.Warnings, warning)
	}

	// 5. Government Warning — strict verbatim check with legibility awareness
	govPresent := res.ExtractedFields.GovernmentWarningFound
	govStatus := models.StatusMissingOnLabel
	govConf := 0
	govExplanation := "Government warning was not detected on the label."
	var govDiff *models.GovWarningComparison
	govSim := 0.0

	var govWarning *models.AIGovernmentWarning
	verbatimText := ""
	if aiCandidates != nil {
		govWarning = &aiCandidates.GovernmentWarning
		verbatimText = govWarning.VerbatimText
		if govWarning.BodyVerbatim != "" {
			verbatimText = govWarning.BodyVerbatim
			if govWarning.PrefixExactCaps {
				verbatimText = "GOVERNMENT WARNING: " + verbatimText
			} else if govWarning.PrefixSeen {
				verbatimText = "Government Warning: " + verbatimText
			}
		}
	}

	if res.ExpectedFields.GovernmentWarningPresent {
		if govWarning != nil && govWarning.Legibility == "illegible" {
			govStatus = models.StatusUncertain
			govConf = 30
			govExplanation = "Government warning area detected but text is illegible. Cannot verify compliance."
		} else if govWarning != nil && govWarning.Legibility == "partial" && govWarning.BodyConfidence < 0.5 {
			govStatus = models.StatusUncertain
			govConf = 50
			govExplanation = "Government warning partially legible. Body text confidence too low for reliable comparison."
		} else if govWarning != nil && govWarning.PrefixSeen && !govWarning.PrefixExactCaps {
			govStatus = models.StatusMismatch
			govConf = 90
			govExplanation = "Government warning prefix is not in required ALL CAPS format per 27 CFR § 16.21."
			res.Warnings = append(res.Warnings, "GOVERNMENT WARNING: prefix must be all uppercase.")
		} else if govPresent && verbatimText != "" {
			comparison := validators.CompareGovernmentWarningVerbatim(verbatimText)
			govDiff = &comparison
			govSim = comparison.Similarity
			if comparison.IsExactMatch {
				govStatus = models.StatusMatch
				govConf = 100
				govExplanation = "Government warning text matches statutory requirement exactly."
			} else if comparison.Similarity < 0.6 {
				govStatus = models.StatusMismatch
				govConf = 90
				govExplanation = "Government warning text appears to be pseudo-text or hallucinated. Character similarity too low."
				res.Warnings = append(res.Warnings, "Government warning body text has very low similarity to statutory requirement. Possible AI-generated or corrupted text.")
			} else {
				govStatus = models.StatusMismatch
				govConf = 85
				govExplanation = fmt.Sprintf("Government warning text differs from statute (%.0f%% similarity). Exact match required per 27 CFR § 16.21.", comparison.Similarity*100)
			}
		} else if govPresent {
			govStatus = models.StatusUncertain
			govConf = 70
			govExplanation = "Government warning prefix detected but body text could not be extracted for strict comparison."
		} else {
			govConf = 90
			govExplanation = "Government warning was expected but not detected on the label."
		}
	} else {
		if govPresent {
			govStatus = models.StatusUncertain
			govConf = 80
			govExplanation = "Government warning found but was not expected in application data."
		} else {
			govStatus = models.StatusMatch
			govConf = 100
			govExplanation = "Government warning correctly omitted."
		}
	}

	govExpStr := "Present and verbatim match required"
	if !res.ExpectedFields.GovernmentWarningPresent {
		govExpStr = "Not required"
	}
	govFoundStr := "Not found"
	if govPresent {
		govFoundStr = "Present"
	}

	if govWarning != nil && govWarning.PrefixSeen && !govWarning.PrefixBold {
		res.Warnings = append(res.Warnings, "GOVERNMENT WARNING: prefix may not be in bold type as required by 27 CFR § 16.22.")
	}

	res.Fields = append(res.Fields, models.FieldCheckResult{
		Field:          "Government Warning",
		Expected:       govExpStr,
		Found:          govFoundStr,
		Status:         govStatus,
		Confidence:     govConf,
		Similarity:     govSim,
		Explanation:    govExplanation,
		Rule:           getRule(catalog, "health_warning_statement"),
		GovWarningDiff: govDiff,
	})

	// 6. Producer/Bottler — fuzzy presence check
	producerFound := res.ExtractedFields.ProducerBottler
	producerAIConf := 0.0
	if aiCandidates != nil {
		producerAIConf = aiCandidates.ProducerOrBottler.Confidence
	}
	producerField := compareFuzzyField("Producer/Bottler", res.ExpectedFields.ProducerBottler, producerFound, producerAIConf, 0.80, 0.60)
	producerField.Rule = getRule(catalog, prefix+"producer_bottler")
	res.Fields = append(res.Fields, producerField)

	// 7. Country of Origin — word overlap matching
	countryFound := res.ExtractedFields.CountryOfOrigin
	countryAIConf := 0.0
	if aiCandidates != nil {
		countryAIConf = aiCandidates.CountryOfOrigin.Confidence
	}
	countryField := compareCountryField(res.ExpectedFields.CountryOfOrigin, countryFound, countryAIConf)
	countryField.Rule = getRule(catalog, prefix+"country_of_origin")
	res.Fields = append(res.Fields, countryField)

	// Overall status
	overall := models.StatusMatch
	confidenceSum := 0
	for _, f := range res.Fields {
		confidenceSum += f.Confidence
		if f.Status == models.StatusMismatch || f.Status == models.StatusMissingOnLabel {
			overall = models.StatusMismatch
		} else if f.Status == models.StatusUncertain && overall != models.StatusMismatch {
			overall = models.StatusUncertain
		}
	}
	if len(res.Fields) > 0 {
		res.OverallConfidence = confidenceSum / len(res.Fields)
	}
	res.OverallStatus = overall

	return res
}

func compareFuzzyField(fieldName, expected, found string, aiConf float64, matchThreshold, uncertainThreshold float64) models.FieldCheckResult {
	if found == "" {
		conf := 0
		expl := fmt.Sprintf("AI did not detect %s on the label.", fieldName)
		if expected != "" {
			return models.FieldCheckResult{
				Field: fieldName, Expected: expected, Found: "Not found",
				Status: models.StatusMissingOnLabel, Confidence: 40, AIConfidence: aiConf, Explanation: expl,
			}
		}
		return models.FieldCheckResult{
			Field: fieldName, Expected: expected, Found: "",
			Status: models.StatusMissingOnLabel, Confidence: conf, AIConfidence: aiConf, Explanation: expl,
		}
	}

	if expected == "" {
		return models.FieldCheckResult{
			Field: fieldName, Expected: expected, Found: found,
			Status: models.StatusMissingInApp, Confidence: 90, Similarity: 0, AIConfidence: aiConf,
			Explanation: fmt.Sprintf("%s found on label but no expected value provided.", fieldName),
		}
	}

	sim := validators.BigramSimilarity(expected, found)
	status := models.StatusMismatch
	conf := int(sim * 100)
	expl := fmt.Sprintf("%s does not match expected value (%.0f%% similarity).", fieldName, sim*100)

	if sim >= matchThreshold {
		status = models.StatusMatch
		conf = 95
		expl = fmt.Sprintf("%s matches expected value (%.0f%% similarity).", fieldName, sim*100)
	} else if sim >= uncertainThreshold {
		status = models.StatusUncertain
		expl = fmt.Sprintf("%s partially matches (%.0f%% similarity). Review recommended.", fieldName, sim*100)
	}

	if aiConf > 0 && aiConf < 0.70 && status == models.StatusMatch {
		status = models.StatusUncertain
		expl = fmt.Sprintf("%s text matches but AI extraction confidence is low (%.0f%%).", fieldName, aiConf*100)
	}

	return models.FieldCheckResult{
		Field: fieldName, Expected: expected, Found: found,
		Status: status, Confidence: conf, Similarity: sim, AIConfidence: aiConf, Explanation: expl,
	}
}

func compareCountryField(expected, found string, aiConf float64) models.FieldCheckResult {
	if found == "" {
		if expected != "" {
			return models.FieldCheckResult{
				Field: "Country of Origin", Expected: expected, Found: "Not found",
				Status: models.StatusMissingOnLabel, Confidence: 40, AIConfidence: aiConf,
				Explanation: "Country of origin not found on label.",
			}
		}
		return models.FieldCheckResult{
			Field: "Country of Origin", Expected: expected, Found: "",
			Status: models.StatusMissingOnLabel, Confidence: 0, AIConfidence: aiConf,
			Explanation: "Country of origin not detected.",
		}
	}
	if expected == "" {
		return models.FieldCheckResult{
			Field: "Country of Origin", Expected: "", Found: found,
			Status: models.StatusMissingInApp, Confidence: 90, AIConfidence: aiConf,
			Explanation: "Country of origin found on label but no expected value provided.",
		}
	}

	overlap := validators.WordOverlap(expected, found)
	sim := validators.BigramSimilarity(expected, found)
	bestSim := overlap
	if sim > bestSim {
		bestSim = sim
	}

	status := models.StatusMismatch
	conf := int(bestSim * 100)
	expl := fmt.Sprintf("Country of origin does not match (%.0f%% similarity).", bestSim*100)

	if bestSim >= 0.50 {
		status = models.StatusMatch
		conf = 90
		expl = fmt.Sprintf("Country of origin matches (%.0f%% similarity).", bestSim*100)
	}

	return models.FieldCheckResult{
		Field: "Country of Origin", Expected: expected, Found: found,
		Status: status, Confidence: conf, Similarity: bestSim, AIConfidence: aiConf, Explanation: expl,
	}
}
