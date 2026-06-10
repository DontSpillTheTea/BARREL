package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

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

		provider := r.FormValue("ocr_provider")
		var expected models.ExpectedLabelFields
		if expectedStr := r.FormValue("expected_json"); expectedStr != "" {
			json.Unmarshal([]byte(expectedStr), &expected)
		}
		beverageType := r.FormValue("beverage_type")
		if beverageType == "" {
			beverageType = "distilled_spirits"
		}

		job := jobStore.CreateJob(header.Filename)

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "accepted",
			"job_id":   job.ID,
			"poll_url": "/api/v1/jobs/" + job.ID,
		})

		// Background processing
		go func(jobID, filename, provider string, beverageType string, expected models.ExpectedLabelFields, fileData []byte) {
			jobStore.UpdateJobStatus(jobID, jobs.StatusProcessing)
			
			inputData := providers.ExtractInput{
				Filename:    filename,
				ContentType: "application/octet-stream",
				Data:        fileData,
			}
			ocrRes, err := ocrManager.Extract(context.Background(), inputData, provider)
			if err != nil {
				res := &models.LabelAnalysisResult{
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
				return
			}

			if ocrRes.Status == "error" {
				res := &models.LabelAnalysisResult{
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
				return
			}

			input := models.AnalysisInput{
				BeverageType:   beverageType,
				Text:           ocrRes.Text,
				ExpectedFields: expected,
			}
			res := analysis.AnalyzeText(input, catalog, ocrRes)
			res.Filename = filename
			
			// Save to storage
			ctx := context.Background()
			_ = storageProvider.SaveImage(ctx, jobID, fileData)
			_ = storageProvider.SaveResult(ctx, jobID, &res)
			
			jobStore.SucceedJob(jobID, &res)
		}(job.ID, header.Filename, provider, beverageType, expected, buf.Bytes())
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
		reviews, err := storageProvider.ListReviews(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(reviews)
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

			review, err := storageProvider.GetReview(r.Context(), jobID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(review)
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
