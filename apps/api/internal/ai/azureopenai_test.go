package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestAzureOpenAIProvider(t *testing.T) {
	provider := NewAzureOpenAIProvider("https://example.openai.azure.com", "dummy-key", "gpt-4o", "2024-05-13")
	provider.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost {
				t.Fatalf("expected POST request, got %s", req.Method)
			}
			if !strings.Contains(req.URL.String(), "/openai/deployments/gpt-4o/chat/completions") {
				t.Fatalf("unexpected request URL: %s", req.URL.String())
			}
			if req.Header.Get("api-key") != "dummy-key" {
				t.Fatalf("expected api-key header to be set")
			}

			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("failed reading request body: %v", err)
			}

			var reqBody map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
				t.Fatalf("failed to decode request JSON: %v", err)
			}

			messages, ok := reqBody["messages"].([]interface{})
			if !ok || len(messages) != 2 {
				t.Fatal("expected messages in request")
			}

			if responseFormat, ok := reqBody["response_format"].(map[string]interface{}); !ok || responseFormat["type"] != "json_object" {
				t.Fatalf("expected json_object response format, got %#v", reqBody["response_format"])
			}

			systemMessage := messages[0].(map[string]interface{})
			if systemMessage["role"] != "system" {
				t.Fatalf("expected first message to be system, got %#v", systemMessage["role"])
			}

			userMessage := messages[1].(map[string]interface{})
			content := userMessage["content"].([]interface{})
			if len(content) != 2 {
				t.Fatalf("expected text and image parts, got %d", len(content))
			}

			imagePart := content[1].(map[string]interface{})
			imageURL := imagePart["image_url"].(map[string]interface{})["url"].(string)
			expectedPrefix := "data:image/png;base64,"
			if !strings.HasPrefix(imageURL, expectedPrefix) {
				t.Fatalf("expected image data URI prefix %q, got %q", expectedPrefix, imageURL)
			}

			encodedImage := strings.TrimPrefix(imageURL, expectedPrefix)
			if _, err := base64.StdEncoding.DecodeString(encodedImage); err != nil {
				t.Fatalf("expected image payload to be valid base64: %v", err)
			}

			respBody := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]interface{}{
							"content": `{"candidates":{"brand_name":{"value":"Test Brand","confidence":0.99,"source":"image"},"class_type":{"value":"Bourbon","confidence":0.98,"source":"image"},"alcohol_content":{"abv":"45% Alc./Vol.","proof":"90 Proof","confidence":0.97,"source":"image"},"net_contents":{"value":"750 mL","confidence":0.96,"source":"image"},"producer_or_bottler":{"value":"Test Producer","confidence":0.9,"source":"image"},"government_warning":{"present":true,"confidence":0.95,"source":"image"},"country_of_origin":{"value":"","confidence":0.1,"source":"image"},"disclosures":{"value":"","confidence":0.1,"source":"image"}}}`,
						},
					},
				},
			}
			respBytes, err := json.Marshal(respBody)
			if err != nil {
				t.Fatalf("failed to encode mock response: %v", err)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(respBytes))),
			}, nil
		}),
	}

	input := SecondReadInput{
		BeverageType: "distilled_spirits",
		ImageBytes:   []byte("dummy-image-data"),
		ContentType:  "image/png",
	}

	result, err := provider.SecondRead(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatalf("Expected result, got nil")
	}
	if result.Candidates.BrandName.Value != "Test Brand" {
		t.Errorf("Expected BrandName 'Test Brand', got '%s'", result.Candidates.BrandName.Value)
	}
	if result.Candidates.ClassType.Value != "Bourbon" {
		t.Errorf("Expected ClassType 'Bourbon', got '%s'", result.Candidates.ClassType.Value)
	}
	if !result.Candidates.GovernmentWarning.Present {
		t.Errorf("Expected GovernmentWarningPresent true, got false")
	}
}
