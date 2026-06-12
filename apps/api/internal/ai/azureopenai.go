package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type AzureOpenAIProvider struct {
	Endpoint   string
	APIKey     string
	Deployment string
	APIVersion string
	HTTPClient *http.Client
}

func NewAzureOpenAIProvider(endpoint, apiKey, deployment, apiVersion string) *AzureOpenAIProvider {
	return &AzureOpenAIProvider{
		Endpoint:   strings.TrimSuffix(endpoint, "/"),
		APIKey:     apiKey,
		Deployment: deployment,
		APIVersion: apiVersion,
		HTTPClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *AzureOpenAIProvider) Name() string {
	return "azure_openai"
}

func (p *AzureOpenAIProvider) SecondRead(ctx context.Context, input SecondReadInput) (*models.AISecondRead, error) {
	if p.Endpoint == "" || p.APIKey == "" || p.Deployment == "" {
		return nil, fmt.Errorf("azure openai is not fully configured")
	}

	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", p.Endpoint, p.Deployment, p.APIVersion)

	promptStr := BuildPrompt(input)
	imageData := input.ImageBytes
	mimeType := normalizeImageContentType(input.ContentType, input.Filename, input.ImageBytes)
	if compressed, err := resizeImage(imageData, mimeType, 768); err == nil && len(compressed) < len(imageData) {
		log.Printf("Compressed image from %d to %d bytes", len(imageData), len(compressed))
		imageData = compressed
		mimeType = "image/jpeg"
	}
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	reqBody := map[string]interface{}{
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "Return only raw valid JSON. Do not include markdown, commentary, or code fences.",
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": promptStr,
					},
					{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url":    dataURI,
							"detail": "low",
						},
					},
				},
			},
		},
		"max_tokens":  1000,
		"temperature": 0.1,
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

	client := p.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("azure openai error: %s (status %d)", string(respBody), resp.StatusCode)
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
		return nil, fmt.Errorf("no response from azure openai")
	}

	content := oaiResp.Choices[0].Message.Content
	cleanContent := CleanJSON(content)

	var result struct {
		Candidates models.AINativeExtraction `json:"candidates"`
	}

	err = json.Unmarshal([]byte(cleanContent), &result)

	// Create the base AISecondRead response
	aiRead := &models.AISecondRead{
		Eligible: true,
		Used:     true,
		Provider: p.Name(),
		RawText:  cleanContent,
		Status:   "success",
	}

	if err != nil {
		aiRead.Status = "error"
		aiRead.Error = fmt.Sprintf("failed to parse AI JSON: %v", err)
	} else {
		aiRead.Candidates = result.Candidates
	}

	return aiRead, nil
}

func normalizeImageContentType(contentType, filename string, imageBytes []byte) string {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if strings.HasPrefix(contentType, "image/") {
		return contentType
	}

	detected := strings.ToLower(http.DetectContentType(imageBytes))
	if strings.HasPrefix(detected, "image/") {
		if semi := strings.Index(detected, ";"); semi >= 0 {
			return detected[:semi]
		}
		return detected
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}

func resizeImage(data []byte, mimeType string, maxDim int) ([]byte, error) {
	var img image.Image
	var err error

	reader := bytes.NewReader(data)
	switch {
	case strings.Contains(mimeType, "png"):
		img, err = png.Decode(reader)
	case strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg"):
		img, err = jpeg.Decode(reader)
	default:
		img, _, err = image.Decode(reader)
	}
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	needsResize := w > maxDim || h > maxDim
	needsCompress := len(data) > 200_000

	if !needsResize && !needsCompress {
		return data, nil
	}

	var dst *image.RGBA
	if needsResize {
		ratio := float64(maxDim) / float64(w)
		if float64(maxDim)/float64(h) < ratio {
			ratio = float64(maxDim) / float64(h)
		}
		newW := int(float64(w) * ratio)
		newH := int(float64(h) * ratio)
		dst = image.NewRGBA(image.Rect(0, 0, newW, newH))
		for y := 0; y < newH; y++ {
			for x := 0; x < newW; x++ {
				srcX := bounds.Min.X + int(float64(x)/ratio)
				srcY := bounds.Min.Y + int(float64(y)/ratio)
				dst.Set(x, y, img.At(srcX, srcY))
			}
		}
	}

	var buf bytes.Buffer
	target := img
	if dst != nil {
		target = dst
	}
	if err := jpeg.Encode(&buf, target, &jpeg.Options{Quality: 50}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
