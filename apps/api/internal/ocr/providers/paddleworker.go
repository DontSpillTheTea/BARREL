package providers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocrclient"
)

type PaddleWorkerProvider struct {
	client *ocrclient.Client
}

func NewPaddleWorkerProvider(workerURL string) *PaddleWorkerProvider {
	return &PaddleWorkerProvider{
		client: ocrclient.New(workerURL),
	}
}

func (p *PaddleWorkerProvider) Name() string {
	return "paddleocr_worker"
}

func (p *PaddleWorkerProvider) Extract(ctx context.Context, input ExtractInput) (*OCRResult, error) {
	reader := bytes.NewReader(input.Data)
	res, err := p.client.Extract(input.Filename, reader, "paddleocr")
	if err != nil {
		return nil, err
	}
	
	if res.Status == "error" {
		return nil, fmt.Errorf("ocr worker error: %s - %s", res.ErrorCode, res.Message)
	}

	return &OCRResult{
		Status:           res.Status,
		Provider:         "paddleocr_worker",
		SelectedProvider: res.SelectedProvider,
		Text:             res.Text,
		MeanConfidence:   res.MeanConfidence,
		Raw:              map[string]any{"raw": res},
		Warnings:         res.Warnings,
	}, nil
}

func (p *PaddleWorkerProvider) Health() (any, error) {
	return p.client.Health()
}

func (p *PaddleWorkerProvider) Ready() (any, error) {
	return p.client.Ready()
}
