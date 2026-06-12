package ocr

import (
	"context"
	"log"
	"os"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocr/providers"
)

type Manager struct {
	primary     providers.Provider
	fallback    providers.Provider
	mock        providers.Provider
	mockGeneric providers.Provider
	useMock     bool
}

func NewManager(workerURL string) *Manager {
	m := &Manager{}

	m.mock = providers.NewMockFastProvider()
	m.mockGeneric = providers.NewMockGenericProvider()
	m.fallback = providers.NewPaddleWorkerProvider(workerURL)

	if os.Getenv("BARREL_OCR_MOCK") == "true" {
		m.useMock = true
		log.Println("OCR mock mode enabled — using generic mock provider for all requests.")
	} else if az, err := providers.NewAzureVisionProvider(); err == nil {
		log.Println("Azure Vision OCR provider configured successfully.")
		m.primary = az
	} else {
		log.Printf("Azure Vision not configured: %v", err)
	}

	return m
}

func (m *Manager) Extract(ctx context.Context, input providers.ExtractInput, requestedProvider string) (*providers.OCRResult, error) {
	if requestedProvider == "mock_fast" {
		return m.mock.Extract(ctx, input)
	}
	if requestedProvider == "mock_generic" || m.useMock {
		return m.mockGeneric.Extract(ctx, input)
	}
	if requestedProvider == "paddleocr_worker" {
		return m.fallback.Extract(ctx, input)
	}
	if requestedProvider == "azure_vision_ocr" && m.primary != nil {
		return m.primary.Extract(ctx, input)
	}

	if m.primary != nil {
		res, err := m.primary.Extract(ctx, input)
		if err == nil {
			return res, nil
		}
		log.Printf("Azure Vision failed: %v, attempting fallback...", err)
	}

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
