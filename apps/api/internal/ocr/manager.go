package ocr

import (
	"context"
	"log"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocr/providers"
)

type Manager struct {
	primary  providers.Provider
	fallback providers.Provider
	mock     providers.Provider
}

func NewManager(workerURL string) *Manager {
	m := &Manager{}
	
	m.mock = providers.NewMockFastProvider()
	m.fallback = providers.NewPaddleWorkerProvider(workerURL)
	
	if az, err := providers.NewAzureVisionProvider(); err == nil {
		log.Println("Azure Vision OCR provider configured successfully.")
		m.primary = az
	} else {
		log.Printf("Azure Vision not configured: %v", err)
	}
	
	return m
}

func (m *Manager) Extract(ctx context.Context, input providers.ExtractInput, requestedProvider string) (*providers.OCRResult, error) {
	// 1. If explicit provider requested (e.g. mock_fast, paddleocr_worker)
	if requestedProvider == "mock_fast" {
		return m.mock.Extract(ctx, input)
	}
	if requestedProvider == "paddleocr_worker" {
		return m.fallback.Extract(ctx, input)
	}
	if requestedProvider == "azure_vision_ocr" && m.primary != nil {
		return m.primary.Extract(ctx, input)
	}

	// 2. Default logic: Try Azure first if available
	if m.primary != nil {
		res, err := m.primary.Extract(ctx, input)
		if err == nil {
			return res, nil
		}
		log.Printf("Azure Vision failed: %v, attempting fallback...", err)
	}

	// 3. Fallback to local
	return m.fallback.Extract(ctx, input)
}

func (m *Manager) WorkerHealth() (any, error) {
	if p, ok := m.fallback.(*providers.PaddleWorkerProvider); ok {
		return p.Health()
	}
	return nil, nil
}

func (m *Manager) WorkerReady() (any, error) {
	if p, ok := m.fallback.(*providers.PaddleWorkerProvider); ok {
		return p.Ready()
	}
	return nil, nil
}
