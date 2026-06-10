package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type AzureVisionProvider struct {
	Endpoint   string
	Key        string
	APIVersion string
	HTTPClient *http.Client
}

func NewAzureVisionProvider() (*AzureVisionProvider, error) {
	endpoint := os.Getenv("AZURE_VISION_ENDPOINT")
	key := os.Getenv("AZURE_VISION_KEY")
	version := os.Getenv("AZURE_VISION_API_VERSION")

	if endpoint == "" || key == "" {
		return nil, errors.New("missing AZURE_VISION_ENDPOINT or AZURE_VISION_KEY")
	}

	if version == "" {
		version = "2024-02-01-preview" // Default fallback for Image Analysis 4.0
	}

	return &AzureVisionProvider{
		Endpoint:   strings.TrimSuffix(endpoint, "/"),
		Key:        key,
		APIVersion: version,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (a *AzureVisionProvider) Name() string {
	return "azure_vision"
}

func (a *AzureVisionProvider) Extract(ctx context.Context, input ExtractInput) (*OCRResult, error) {
	// Let's use Image Analysis 4.0 endpoint: /computervision/imageanalysis:analyze?features=read&api-version=...
	url := fmt.Sprintf("%s/computervision/imageanalysis:analyze?features=read&api-version=%s", a.Endpoint, a.APIVersion)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(input.Data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", a.Key)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("azure vision network error: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("azure vision api error: HTTP %d", resp.StatusCode)
	}

	var raw map[string]any
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse azure response: %v", err)
	}

	// Extract text from Image Analysis 4.0 'readResult'
	var textBuilder strings.Builder
	var totalConfidence float64
	var wordCount int

	if readResult, ok := raw["readResult"].(map[string]any); ok {
		if blocks, ok := readResult["blocks"].([]any); ok {
			for _, blockAny := range blocks {
				block := blockAny.(map[string]any)
				if lines, ok := block["lines"].([]any); ok {
					for _, lineAny := range lines {
						line := lineAny.(map[string]any)
						if text, ok := line["text"].(string); ok {
							textBuilder.WriteString(text + "\n")
						}
						// Collect confidence from words
						if words, ok := line["words"].([]any); ok {
							for _, wordAny := range words {
								word := wordAny.(map[string]any)
								if conf, ok := word["confidence"].(float64); ok {
									totalConfidence += conf
									wordCount++
								}
							}
						}
					}
				}
			}
		}
	}

	meanConf := 0.0
	if wordCount > 0 {
		meanConf = (totalConfidence / float64(wordCount)) * 100.0 // Convert to percentage 0-100
	}

	return &OCRResult{
		Status:           "ok",
		Provider:         "azure_vision",
		SelectedProvider: "azure_vision",
		Text:             strings.TrimSpace(textBuilder.String()),
		MeanConfidence:   meanConf,
		Raw:              raw,
		Warnings:         nil,
	}, nil
}
