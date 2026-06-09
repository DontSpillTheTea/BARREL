package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/analysis"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/config"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocrclient"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/rules"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/security"
)

type APIResponse struct {
	Status   string                  `json:"status"`
	Service  string                  `json:"service"`
	Filename string                  `json:"filename,omitempty"`
	OCR      *ocrclient.OCRResponse  `json:"ocr,omitempty"`
	Error    string                  `json:"error,omitempty"`
}

func main() {
	cfg := config.Load()
	ocrClient := ocrclient.New(cfg.OCRWorkerURL)
	catalog, err := rules.LoadCatalog(cfg.RulesPath)
	if err != nil {
		log.Printf("Warning: Failed to load rules catalog: %v", err)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "barrel-api"})
	})

	http.HandleFunc("/health/ocr-worker", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res, err := ocrClient.Health()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "service": "barrel-api", "error": "ocr_worker_unreachable"})
			return
		}
		json.NewEncoder(w).Encode(res)
	})

	http.HandleFunc("/api/v1/ocr/extract", func(w http.ResponseWriter, r *http.Request) {
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
		res, err := ocrClient.Extract(header.Filename, file, provider)
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
	})

	http.HandleFunc("/api/v1/labels/analyze-text", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.HandleFunc("/api/v1/labels/analyze", func(w http.ResponseWriter, r *http.Request) {
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
		ocrRes, err := ocrClient.Extract(header.Filename, file, provider)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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

		input := models.AnalysisInput{
			BeverageType:   beverageType,
			Text:           ocrRes.Text,
			ExpectedFields: expected,
		}

		res := analysis.AnalyzeText(input, catalog, ocrRes)
		res.Filename = header.Filename
		json.NewEncoder(w).Encode(res)
	})

	port := ":8080"
	log.Printf("Starting Go API on port %s...\n", port)
	handler := security.CORS(http.DefaultServeMux)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
