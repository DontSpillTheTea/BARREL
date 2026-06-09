package ocrclient

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type ImageQuality struct {
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	ContrastScore float64  `json:"contrast_score"`
	BlurScore     float64  `json:"blur_score"`
	QualityNotes  []string `json:"quality_notes"`
}

type OCRResponse struct {
	Status         string       `json:"status"`
	Service        string       `json:"service"`
	ErrorCode      string       `json:"error_code,omitempty"`
	Message        string       `json:"message,omitempty"`
	Filename       string       `json:"filename,omitempty"`
	ContentType    string       `json:"content_type,omitempty"`
	OCREngine      string       `json:"ocr_engine,omitempty"`
	Text           string       `json:"text,omitempty"`
	MeanConfidence float64      `json:"mean_confidence,omitempty"`
	ImageQuality   ImageQuality `json:"image_quality,omitempty"`
	Warnings       []string     `json:"warnings,omitempty"`
}

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Health() (*OCRResponse, error) {
	resp, err := c.http.Get(c.baseURL + "/health")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out OCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Extract(filename string, file io.Reader) (*OCRResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/ocr/extract", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out OCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
