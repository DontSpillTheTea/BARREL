package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type AzureOpenAIProvider struct {
	Endpoint   string
	APIKey     string
	Deployment string
	APIVersion string
}

func NewAzureOpenAIProvider(endpoint, apiKey, deployment, apiVersion string) *AzureOpenAIProvider {
	return &AzureOpenAIProvider{
		Endpoint:   strings.TrimSuffix(endpoint, "/"),
		APIKey:     apiKey,
		Deployment: deployment,
		APIVersion: apiVersion,
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
	base64Image := base64.StdEncoding.EncodeToString(input.ImageBytes)
	mimeType := input.ContentType
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	reqBody := map[string]interface{}{
		"messages": []map[string]interface{}{
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
							"url": dataURI,
						},
					},
				},
			},
		},
		"max_tokens":  2000,
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

	client := &http.Client{Timeout: 45 * time.Second}
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
		Candidates models.AISecondReadCandidates `json:"candidates"`
		Findings   []models.AISecondReadFinding  `json:"findings"`
	}

	err = json.Unmarshal([]byte(cleanContent), &result)
	
	// Create the base AISecondRead response
	aiRead := &models.AISecondRead{
		Eligible:   true,
		Used:       true,
		Provider:   p.Name(),
		RawText:    cleanContent,
		Status:     "success",
	}

	if err != nil {
		aiRead.Status = "error"
		aiRead.Error = fmt.Sprintf("failed to parse AI JSON: %v", err)
	} else {
		aiRead.Candidates = result.Candidates
		aiRead.Findings = result.Findings
	}

	return aiRead, nil
}
