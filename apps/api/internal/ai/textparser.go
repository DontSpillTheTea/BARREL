package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type TextParserProvider struct {
	Endpoint   string
	APIKey     string
	Deployment string
	APIVersion string
	HTTPClient *http.Client
}

func NewTextParserProvider(endpoint, apiKey, deployment, apiVersion string) *TextParserProvider {
	if deployment == "" {
		deployment = "barrel-ai-native-parser"
	}
	return &TextParserProvider{
		Endpoint:   strings.TrimSuffix(endpoint, "/"),
		APIKey:     apiKey,
		Deployment: deployment,
		APIVersion: apiVersion,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func BuildTextParserPrompt(ocrText, beverageType string, expected models.ExpectedLabelFields) string {
	expectedJSON, _ := json.MarshalIndent(expected, "", "  ")

	return fmt.Sprintf(`You are extracting structured fields from OCR text of a beverage alcohol label.
The beverage type is: %s
Do not decide legal compliance. Do not invent missing values.
If a field is not clearly present in the OCR text, set its value to null or empty with low confidence.
For the government_warning field: transcribe the EXACT text as it appears. Do not correct or normalize it.
Return only raw valid JSON. No markdown or code blocks.

Expected Fields for reference (do not copy them; extract what the OCR text actually contains):
%s

OCR Text:
%s

Return exactly this JSON structure:
{
  "candidates": {
    "brand_name": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "ocr_text" },
    "class_type": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "ocr_text" },
    "alcohol_content": { "abv": "...", "proof": "...", "confidence": 0.99, "evidence": "...", "source": "ocr_text" },
    "net_contents": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "ocr_text" },
    "producer_or_bottler": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "ocr_text" },
    "government_warning": { "present": true, "verbatim_text": "...", "confidence": 0.99, "source": "ocr_text", "possible_typos": [] },
    "country_of_origin": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "ocr_text" },
    "disclosures": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "ocr_text" },
    "image_quality_flags": []
  }
}`, beverageType, string(expectedJSON), ocrText)
}

func (p *TextParserProvider) ParseText(ctx context.Context, ocrText, beverageType string, expected models.ExpectedLabelFields) (*models.AINativeExtraction, error) {
	if p.Endpoint == "" || p.APIKey == "" || p.Deployment == "" {
		return nil, fmt.Errorf("text parser provider not configured")
	}

	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", p.Endpoint, p.Deployment, p.APIVersion)

	prompt := BuildTextParserPrompt(ocrText, beverageType, expected)

	reqBody := map[string]interface{}{
		"response_format": map[string]interface{}{"type": "json_object"},
		"messages": []map[string]interface{}{
			{"role": "system", "content": "Return only raw valid JSON. No markdown, commentary, or code fences."},
			{"role": "user", "content": prompt},
		},
		"max_tokens":  800,
		"temperature": 0.0,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", p.APIKey)

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("text parser error: %s (status %d)", string(respBody), resp.StatusCode)
	}

	var oaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return nil, err
	}
	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from text parser")
	}

	content := CleanJSON(oaiResp.Choices[0].Message.Content)
	var result struct {
		Candidates models.AINativeExtraction `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse text parser JSON: %v", err)
	}

	return &result.Candidates, nil
}
