package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/config"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocrclient"
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

		res, err := ocrClient.Extract(header.Filename, file)
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

	port := ":8080"
	log.Printf("Starting Go API on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
