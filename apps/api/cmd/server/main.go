package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/ai"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/analysis"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/config"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/jobs"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocr"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocr/providers"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/rules"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/security"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/storage"
)

type APIResponse struct {
	Status   string               `json:"status"`
	Service  string               `json:"service"`
	Filename string               `json:"filename,omitempty"`
	OCR      *providers.OCRResult `json:"ocr,omitempty"`
	Error    string               `json:"error,omitempty"`
}

func main() {
	cfg := config.Load()
	ocrManager := ocr.NewManager(cfg.OCRWorkerURL)
	catalog, err := rules.LoadCatalog(cfg.RulesPath)
	if err != nil {
		log.Printf("Warning: Failed to load rules catalog: %v", err)
	}

	var aiProvider ai.Provider
	if cfg.AzureOpenAIEndpoint != "" && cfg.AzureOpenAIAPIKey != "" && cfg.AzureOpenAIAPIKey != "dummy" {
		aiProvider = ai.NewAzureOpenAIProvider(cfg.AzureOpenAIEndpoint, cfg.AzureOpenAIAPIKey, cfg.AzureOpenAIDeployment, cfg.AzureOpenAIAPIVersion)
		log.Println("Azure OpenAI provider configured.")
	} else {
		aiProvider = ai.NewMockProvider()
		log.Println("Azure OpenAI credentials missing or dummy; using Mock AI provider.")
	}

	jobStore := jobs.NewStore()
	storageProvider := storage.NewProvider()
	log.Printf("Using storage provider: %T", storageProvider)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "barrel-api"})
	})

	security.RegisterAuthRoutes()

	http.HandleFunc("/api/v1/ocr/extract", security.RequireToken(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(APIResponse{Status: "error", Service: "barrel-api", Error: "method_not_allowed"})
			return
		}

		if err := security.ValidateUpload(r, cfg.MaxUploadMB); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIResponse{Status: "error", Service: "barrel-api", Error: "upload_too_large_or_invalid"})
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIResponse{Status: "error", Service: "barrel-api", Error: "missing_file"})
			return
		}
		defer file.Close()

		if !security.IsAllowedExtension(header.Filename) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIResponse{Status: "error", Service: "barrel-api", Error: "unsupported_file_extension"})
			return
		}

		provider := r.FormValue("ocr_provider")
		if provider == "" {
			provider = "azure_vision_ocr"
		}

		buf := &bytes.Buffer{}
		io.Copy(buf, file)
		input := providers.ExtractInput{
			Filename:    header.Filename,
			ContentType: header.Header.Get("Content-Type"),
			Data:        buf.Bytes(),
		}

		res, err := ocrManager.Extract(r.Context(), input, provider)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIResponse{Status: "error", Service: "barrel-api", Error: "ocr_worker_communication_error"})
			return
		}

		status := "ok"
		if res.Status == "error" {
			status = "error"
		}

		json.NewEncoder(w).Encode(APIResponse{
			Status:   status,
			Service:  "barrel-api",
			Filename: header.Filename,
			OCR:      res,
		})
	}))

	http.HandleFunc("/api/v1/labels/analyze-text", security.RequireToken(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var input models.AnalysisInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
			return
		}

		res := analysis.AnalyzeText(input, catalog, nil)
		json.NewEncoder(w).Encode(res)
	}))

	http.HandleFunc("/api/v1/labels/analyze", security.RequireToken(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := security.ValidateUpload(r, cfg.MaxUploadMB); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		if !security.IsAllowedExtension(header.Filename) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		provider := r.FormValue("ocr_provider")
		if provider == "" {
			provider = "azure_vision_ocr"
		}

		buf := &bytes.Buffer{}
		io.Copy(buf, file)
		input := providers.ExtractInput{
			Filename:    header.Filename,
			ContentType: header.Header.Get("Content-Type"),
			Data:        buf.Bytes(),
		}

		ocrRes, err := ocrManager.Extract(r.Context(), input, provider)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIResponse{Status: "error", Service: "barrel-api", Error: "ocr_worker_communication_error"})
			return
		}

		if ocrRes.Status == "error" {
			// Structured OCR error
			res := models.LabelAnalysisResult{
				OverallStatus:     "Needs Review",
				OverallConfidence: 0,
				OCR:               ocrRes,
				AIEscalation: models.AIEscalation{
					Eligible: true,
					Used:     false,
					Provider: "none",
					Reason:   "Local OCR provider was not ready or failed.",
				},
			}
			w.WriteHeader(http.StatusOK) // Return 200 with structured Needs Review, as requested
			json.NewEncoder(w).Encode(res)
			return
		}

		var expected models.ExpectedLabelFields
		if expectedStr := r.FormValue("expected_json"); expectedStr != "" {
			json.Unmarshal([]byte(expectedStr), &expected)
		}
		beverageType := r.FormValue("beverage_type")
		if beverageType == "" {
			beverageType = "distilled_spirits"
		}

		analysisInput := models.AnalysisInput{
			BeverageType:   beverageType,
			Text:           ocrRes.Text,
			ExpectedFields: expected,
		}

		res := analysis.AnalyzeText(analysisInput, catalog, ocrRes)
		res.Filename = header.Filename
		json.NewEncoder(w).Encode(res)
	}))

	http.HandleFunc("/api/v1/labels/analyze-async", security.RequireToken(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := security.ValidateUpload(r, cfg.MaxUploadMB); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !security.IsAllowedExtension(header.Filename) {
			file.Close()
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Read file into memory safely
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, file); err != nil {
			file.Close()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))

		provider := r.FormValue("ocr_provider")
		if provider == "" {
			provider = "ai_native"
		}
		var expected models.ExpectedLabelFields
		if expectedStr := r.FormValue("expected_json"); expectedStr != "" {
			json.Unmarshal([]byte(expectedStr), &expected)
		}
		beverageType := r.FormValue("beverage_type")
		if beverageType == "" {
			beverageType = "distilled_spirits"
		}

		if ext == ".zip" {
			zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var spawnedJobs []map[string]string
			for _, zf := range zipReader.File {
				zExt := strings.ToLower(filepath.Ext(zf.Name))
				if zExt == ".png" || zExt == ".jpg" || zExt == ".jpeg" {
					rc, err := zf.Open()
					if err != nil {
						continue
					}
					var zBuf bytes.Buffer
					io.Copy(&zBuf, rc)
					rc.Close()

					job := jobStore.CreateJob(zf.Name)
					spawnedJobs = append(spawnedJobs, map[string]string{
						"job_id":   job.ID,
						"filename": zf.Name,
					})

					go processJob(job.ID, zf.Name, provider, beverageType, expected, zBuf.Bytes(), jobStore, ocrManager, catalog, aiProvider, cfg, storageProvider)
				}
			}
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "accepted",
				"batch":  true,
				"jobs":   spawnedJobs,
			})
			return
		}

		job := jobStore.CreateJob(header.Filename)

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "accepted",
			"batch":    false,
			"job_id":   job.ID,
			"poll_url": "/api/v1/jobs/" + job.ID,
		})

		go processJob(job.ID, header.Filename, provider, beverageType, expected, buf.Bytes(), jobStore, ocrManager, catalog, aiProvider, cfg, storageProvider)
	}))

	http.HandleFunc("/api/v1/jobs/", security.RequireToken(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		jobID := r.URL.Path[len("/api/v1/jobs/"):]
		if jobID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		job, ok := jobStore.GetJob(jobID)
		if !ok {
			record, err := storageProvider.GetReview(r.Context(), jobID)
			if err == nil {
				log.Printf("jobs route fallback hit for %s via storage", jobID)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "succeeded",
					"result": record.Result,
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("jobs route in-memory hit for %s status=%s", jobID, job.Status)

		json.NewEncoder(w).Encode(job)
	}))

	http.HandleFunc("/api/v1/reviews", security.RequireToken(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		records, err := storageProvider.ListReviews(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		liveJobs := jobStore.ListJobs()
		log.Printf("reviews route: persisted=%d live_jobs=%d", len(records), len(liveJobs))

		summariesByID := map[string]models.ReviewSummary{}
		for _, rec := range records {
			summary := recordToSummary(rec)
			summariesByID[summary.JobID] = summary
		}

		for _, job := range liveJobs {
			summary := summaryFromJob(job)
			if existing, ok := summariesByID[summary.JobID]; ok {
				if summary.ReviewerDecision == "" {
					summary.ReviewerDecision = existing.ReviewerDecision
				}
			}
			summariesByID[summary.JobID] = summary
		}

		summaries := make([]models.ReviewSummary, 0, len(summariesByID))
		for _, summary := range summariesByID {
			summaries = append(summaries, summary)
		}
		sort.Slice(summaries, func(i, j int) bool { return summaries[i].SubmittedAt > summaries[j].SubmittedAt })

		json.NewEncoder(w).Encode(map[string]interface{}{
			"reviews": summaries,
		})
	}))

	http.HandleFunc("/api/v1/reviews/", security.RequireToken(func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Path[len("/api/v1/reviews/"):]
		if jobID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {
			if len(jobID) > 6 && jobID[len(jobID)-6:] == "/image" {
				id := jobID[:len(jobID)-6]
				img, err := storageProvider.GetImage(r.Context(), id)
				if err != nil {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "image/png")
				w.Write(img)
				return
			}

			record, err := storageProvider.GetReview(r.Context(), jobID)
			if err == nil {
				detail := recordToDetail(record)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(detail)
				return
			}
			if job, ok := jobStore.GetJob(jobID); ok && job.Result != nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(detailFromJob(job))
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method == http.MethodPost {
			if len(jobID) > 9 && jobID[len(jobID)-9:] == "/decision" {
				id := jobID[:len(jobID)-9]
				var dec storage.ReviewDecision
				if err := json.NewDecoder(r.Body).Decode(&dec); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if err := storageProvider.SaveDecision(r.Context(), id, dec); err != nil {
					if !jobStore.SetDecision(id, dec.Decision, dec.Notes) {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else {
					jobStore.SetDecision(id, dec.Decision, dec.Notes)
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
				return
			}
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	log.Printf("Starting Go API on port %s...\n", port)
	handler := security.CORS(http.DefaultServeMux)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func recordToSummary(r storage.ReviewRecord) models.ReviewSummary {
	sum := models.ReviewSummary{
		ID:               r.JobID,
		JobID:            r.JobID,
		Filename:         r.Filename,
		SubmittedAt:      r.Timestamp,
		ReviewerDecision: r.Status,
	}
	if r.Result != nil {
		if sum.Filename == "" {
			sum.Filename = r.Result.Filename
		}
		sum.OverallStatus = r.Result.OverallStatus
		sum.OverallConfidence = r.Result.OverallConfidence
		sum.BeverageType = r.Result.BeverageType

		sum.ProviderRequested = r.Result.RequestedProvider

		providerUsed := "unknown"
		if r.Result.AIEscalation.Used {
			providerUsed = r.Result.AIEscalation.Provider
		} else if r.Result.OCR != nil {
			providerUsed = r.Result.OCR.SelectedProvider
		} else if r.Result.RequestedProvider != "" {
			providerUsed = r.Result.RequestedProvider
		}
		sum.ProviderUsed = providerUsed

		sum.BrandName = r.Result.ExtractedFields.BrandName
		sum.ClassType = r.Result.ExtractedFields.ClassType
		sum.AlcoholContent = r.Result.ExtractedFields.AlcoholContent
		sum.NetContents = r.Result.ExtractedFields.NetContents

		passCount := 0
		for _, f := range r.Result.Fields {
			if f.Status == models.StatusMatch || f.Status == "Pass" {
				passCount++
			}
		}
		sum.FieldPassCount = passCount
		sum.FieldTotalCount = len(r.Result.Fields)
	}
	return sum
}

func recordToDetail(r *storage.ReviewRecord) models.ReviewDetail {
	det := models.ReviewDetail{
		Summary: recordToSummary(*r),
	}
	if r.Result != nil {
		det.Result = *r.Result
		if r.Result.OCR != nil {
			det.RawOCRText = r.Result.OCR.Text
		}
	}
	if r.HasImage {
		det.OriginalImageURL = "/api/v1/reviews/" + r.JobID + "/image"
	}
	return det
}

func summaryFromJob(job *jobs.Job) models.ReviewSummary {
	sum := models.ReviewSummary{
		ID:               job.ID,
		JobID:            job.ID,
		Filename:         job.Filename,
		SubmittedAt:      job.CreatedAt.Format(time.RFC3339),
		CompletedAt:      job.UpdatedAt.Format(time.RFC3339),
		ReviewerDecision: job.Decision,
	}
	if job.Result != nil {
		sum.OverallStatus = job.Result.OverallStatus
		sum.OverallConfidence = job.Result.OverallConfidence
		sum.BeverageType = job.Result.BeverageType
		sum.ProviderRequested = job.Result.RequestedProvider
		sum.BrandName = job.Result.ExtractedFields.BrandName
		sum.ClassType = job.Result.ExtractedFields.ClassType
		sum.AlcoholContent = job.Result.ExtractedFields.AlcoholContent
		sum.NetContents = job.Result.ExtractedFields.NetContents

		providerUsed := "unknown"
		if job.Result.AIEscalation.Used {
			providerUsed = job.Result.AIEscalation.Provider
		} else if job.Result.OCR != nil {
			providerUsed = job.Result.OCR.SelectedProvider
		} else if job.Result.RequestedProvider != "" {
			providerUsed = job.Result.RequestedProvider
		}
		sum.ProviderUsed = providerUsed

		passCount := 0
		for _, f := range job.Result.Fields {
			if f.Status == models.StatusMatch || f.Status == "Pass" {
				passCount++
			}
		}
		sum.FieldPassCount = passCount
		sum.FieldTotalCount = len(job.Result.Fields)
	}
	return sum
}

func detailFromJob(job *jobs.Job) models.ReviewDetail {
	detail := models.ReviewDetail{
		Summary: summaryFromJob(job),
	}
	if job.Result != nil {
		detail.Result = *job.Result
		if job.Result.OCR != nil {
			detail.RawOCRText = job.Result.OCR.Text
		}
	}
	return detail
}

func processJob(jobID, filename, provider string, beverageType string, expected models.ExpectedLabelFields, fileData []byte, jobStore *jobs.Store, ocrManager *ocr.Manager, catalog *rules.Catalog, aiProvider ai.Provider, cfg config.Config, storageProvider storage.Provider) {
	startTime := time.Now()
	jobStore.UpdateJobStatus(jobID, jobs.StatusProcessing)
	log.Printf("Processing job %s with requested provider=%s aiProviderType=%T storageProviderType=%T", jobID, provider, aiProvider, storageProvider)

	inputData := providers.ExtractInput{
		Filename:    filename,
		ContentType: "application/octet-stream",
		Data:        fileData,
	}
	ctx := context.Background()
	if err := storageProvider.SaveImage(ctx, jobID, fileData); err != nil {
		log.Printf("failed to save image for job %s: %v", jobID, err)
	}

	if provider == "ai_native" && !cfg.AINativeEnabled {
		res := models.LabelAnalysisResult{
			Filename:          filename,
			RequestedProvider: provider,
			BeverageType:      beverageType,
			OverallStatus:     "Error",
			ExpectedFields:    expected,
			AIEscalation: models.AIEscalation{
				Eligible: true,
				Used:     false,
				Provider: "ai_native",
				Reason:   "AI-native parser is disabled",
			},
		}
		jobStore.FailJob(jobID, "ai_native_disabled", &res)
		_ = storageProvider.SaveResult(ctx, jobID, &res)
		return
	}

	if provider == "ai_native" && aiProvider == nil {
		res := models.LabelAnalysisResult{
			Filename:          filename,
			RequestedProvider: provider,
			BeverageType:      beverageType,
			OverallStatus:     "Error",
			ExpectedFields:    expected,
			AIEscalation: models.AIEscalation{
				Eligible: true,
				Used:     false,
				Provider: "ai_native",
				Reason:   "AI provider not configured",
			},
		}
		jobStore.FailJob(jobID, "ai_provider_not_configured", &res)
		_ = storageProvider.SaveResult(ctx, jobID, &res)
		return
	}
	var ocrRes *providers.OCRResult
	var err error
	if provider != "ai_native" {
		ocrRes, err = ocrManager.Extract(ctx, inputData, provider)
		if err != nil {
			res := &models.LabelAnalysisResult{
				Filename:          filename,
				RequestedProvider: provider,
				OverallStatus:     "Needs Review",
				OverallConfidence: 0,
				AIEscalation: models.AIEscalation{
					Eligible: true,
					Used:     false,
					Provider: "none",
					Reason:   "Accurate local OCR failed or timed out.",
				},
			}
			jobStore.FailJob(jobID, "ocr_worker_communication_error", res)
			_ = storageProvider.SaveResult(ctx, jobID, res)
			return
		}

		if ocrRes.Status == "error" {
			res := &models.LabelAnalysisResult{
				Filename:          filename,
				RequestedProvider: provider,
				OverallStatus:     "Needs Review",
				OverallConfidence: 0,
				OCR:               ocrRes,
				AIEscalation: models.AIEscalation{
					Eligible: true,
					Used:     false,
					Provider: "none",
					Reason:   "Local OCR provider was not ready or failed.",
				},
			}
			jobStore.FailJob(jobID, "ocr_provider_error", res)
			_ = storageProvider.SaveResult(ctx, jobID, res)
			return
		}
	}

	var res models.LabelAnalysisResult

	// Get OCR data if requested
	if provider != "ai_native" {
		res = analysis.AnalyzeText(models.AnalysisInput{
			BeverageType:   beverageType,
			Text:           ocrRes.Text,
			ExpectedFields: expected,
		}, catalog, ocrRes)
		res.RequestedProvider = provider
		res.OCRText = ocrRes.Text
	} else {
		// Just initialize empty result for AI to populate
		res = models.LabelAnalysisResult{
			BeverageType:      beverageType,
			RequestedProvider: provider,
			ExpectedFields:    expected,
		}
	}

	res.Filename = filename

	// Run the AI-native parser only when ai_native was requested.
	if provider == "ai_native" && aiProvider != nil {
		parserInput := ai.SecondReadInput{
			Filename:       filename,
			ContentType:    http.DetectContentType(fileData),
			ImageBytes:     fileData,
			OCRText:        res.OCRText,
			ExpectedFields: expected,
			BeverageType:   beverageType,
			InitialResult:  res,
		}
		aiRes, err := aiProvider.SecondRead(ctx, parserInput)
		if err != nil {
			res.AIEscalation = models.AIEscalation{
				Eligible: true,
				Used:     false,
				Provider: "ai_native",
				Reason:   err.Error(),
			}
			res.Warnings = append(res.Warnings, "AI-native parser request failed.")
		} else if aiRes != nil {
			res.AISecondRead = aiRes
			res.AIEscalation = models.AIEscalation{
				Eligible: true,
				Used:     true,
				Provider: aiRes.Provider,
				Reason:   aiRes.Reason,
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
			res = analysis.AnalyzeAI(res, catalog)
		}
	}

	res.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	if err := storageProvider.SaveResult(ctx, jobID, &res); err != nil {
		log.Printf("failed to save result for job %s: %v", jobID, err)
	}

	jobStore.SucceedJob(jobID, &res)
}
