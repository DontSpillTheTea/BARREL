package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

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
	Status   string                  `json:"status"`
	Service  string                  `json:"service"`
	Filename string                  `json:"filename,omitempty"`
	OCR      *providers.OCRResult    `json:"ocr,omitempty"`
	Error    string                  `json:"error,omitempty"`
}

func main() {
	cfg := config.Load()
	ocrManager := ocr.NewManager(cfg.OCRWorkerURL)
	catalog, err := rules.LoadCatalog(cfg.RulesPath)
	if err != nil {
		log.Printf("Warning: Failed to load rules catalog: %v", err)
	}

	var aiProvider ai.Provider
	if cfg.AISecondReadEnabled {
		if cfg.AzureOpenAIEndpoint != "" && cfg.AzureOpenAIAPIKey != "" {
			aiProvider = ai.NewAzureOpenAIProvider(cfg.AzureOpenAIEndpoint, cfg.AzureOpenAIAPIKey, cfg.AzureOpenAIDeployment, cfg.AzureOpenAIAPIVersion)
			log.Println("Azure OpenAI provider configured for Second Read.")
		} else {
			aiProvider = ai.NewMockProvider()
			log.Println("Azure OpenAI credentials missing; using Mock AI provider for Second Read.")
		}
	}

	jobStore := jobs.NewStore()
	storageProvider := storage.NewProvider()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "barrel-api"})
	})

	http.HandleFunc("/health/ocr-worker", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res, err := ocrManager.WorkerHealth()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "service": "barrel-api", "error": "ocr_worker_unreachable"})
			return
		}
		json.NewEncoder(w).Encode(res)
	})

	http.HandleFunc("/health/ocr-worker-ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res, err := ocrManager.WorkerReady()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "service": "barrel-api", "error": "ocr_worker_unreachable"})
			return
		}
		json.NewEncoder(w).Encode(res)
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
						"job_id": job.ID,
						"filename": zf.Name,
					})
					
					go processJob(job.ID, zf.Name, provider, beverageType, expected, zBuf.Bytes(), jobStore, ocrManager, catalog, aiProvider, cfg, storageProvider)
				}
			}
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "accepted",
				"batch": true,
				"jobs": spawnedJobs,
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
			w.WriteHeader(http.StatusNotFound)
			return
		}

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
		
		var summaries []models.ReviewSummary
		for _, rec := range records {
			summaries = append(summaries, recordToSummary(rec))
		}
		
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
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			detail := recordToDetail(record)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(detail)
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
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
				return
			}
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	}))

	port := ":8080"
	log.Printf("Starting Go API on port %s...\n", port)
	handler := security.CORS(http.DefaultServeMux)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func recordToSummary(r storage.ReviewRecord) models.ReviewSummary {
	sum := models.ReviewSummary{
		ID:                r.JobID,
		JobID:             r.JobID,
		Filename:          r.Filename,
		SubmittedAt:       r.Timestamp,
		ReviewerDecision:  r.Status,
	}
	if r.Result != nil {
		if sum.Filename == "" {
			sum.Filename = r.Result.Filename
		}
		sum.OverallStatus = r.Result.OverallStatus
		sum.OverallConfidence = r.Result.OverallConfidence
		sum.BeverageType = r.Result.BeverageType
		
		provider := "unknown"
		if r.Result.RequestedProvider != "" {
			provider = r.Result.RequestedProvider
		} else if r.Result.OCR != nil {
			provider = r.Result.OCR.SelectedProvider
		} else if r.Result.AISecondRead != nil && r.Result.AISecondRead.Used {
			provider = r.Result.AISecondRead.Provider
		}
		sum.OCRProvider = provider

		passCount := 0
		for _, f := range r.Result.Fields {
			if f.Status == "Pass" {
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

func processJob(jobID, filename, provider string, beverageType string, expected models.ExpectedLabelFields, fileData []byte, jobStore *jobs.Store, ocrManager *ocr.Manager, catalog *rules.Catalog, aiProvider ai.Provider, cfg config.Config, storageProvider storage.Provider) {
	jobStore.UpdateJobStatus(jobID, jobs.StatusProcessing)
	
	inputData := providers.ExtractInput{
		Filename:    filename,
		ContentType: "application/octet-stream",
		Data:        fileData,
	}
	ctx := context.Background()
	_ = storageProvider.SaveImage(ctx, jobID, fileData)

	if provider == "ai_based" && aiProvider == nil {
		res := &models.LabelAnalysisResult{
			Filename:          filename,
			RequestedProvider: provider,
			OverallStatus:     "Error",
			OverallConfidence: 0,
			AIEscalation: models.AIEscalation{
				Eligible: false,
				Used:     false,
				Provider: "ai_based",
				Reason:   "AI provider not configured",
			},
		}
		jobStore.FailJob(jobID, "ai_provider_not_configured", res)
		_ = storageProvider.SaveResult(ctx, jobID, res)
		return
	}

	ocrRes, err := ocrManager.Extract(ctx, inputData, provider)
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

	input := models.AnalysisInput{
		BeverageType:   beverageType,
		Text:           ocrRes.Text,
		ExpectedFields: expected,
	}
	res := analysis.AnalyzeText(input, catalog, ocrRes)
	res.Filename = filename
	res.RequestedProvider = provider
	
	// Auto AI Second Read or forced if provider is ai_based
	forceAI := provider == "ai_based" || provider == "azure_openai"
	if (forceAI || (cfg.AISecondReadAutoOnFail && res.AISecondRead != nil && res.AISecondRead.Eligible)) && aiProvider != nil {
		secondReadInput := ai.SecondReadInput{
			Filename:       filename,
			ContentType:    "application/octet-stream",
			ImageBytes:     fileData,
			OCRText:        res.OCRText,
			ExpectedFields: expected,
			BeverageType:   beverageType,
			InitialResult:  res,
		}
		aiRes, _ := aiProvider.SecondRead(ctx, secondReadInput)
		if aiRes != nil {
			res.AISecondRead = aiRes
		}
	}

	// Save to storage
	_ = storageProvider.SaveResult(ctx, jobID, &res)
	
	jobStore.SucceedJob(jobID, &res)
}
