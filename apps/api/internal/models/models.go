package models

import "github.com/DontSpillTheTea/barrel/apps/api/internal/ocrclient"

type ExpectedLabelFields struct {
	BrandName                string `json:"brand_name,omitempty"`
	ClassType                string `json:"class_type,omitempty"`
	AlcoholContent           string `json:"alcohol_content,omitempty"`
	NetContents              string `json:"net_contents,omitempty"`
	GovernmentWarningPresent bool   `json:"government_warning_present,omitempty"`
}

type ExtractedLabelFields struct {
	BrandName              string `json:"brand_name,omitempty"`
	ClassType              string `json:"class_type,omitempty"`
	AlcoholContent         string `json:"alcohol_content,omitempty"`
	NetContents            string `json:"net_contents,omitempty"`
	GovernmentWarningFound bool   `json:"government_warning_found,omitempty"`
}

type RuleBreadcrumb struct {
	ID        string `json:"id"`
	Citation  string `json:"citation"`
	SourceURL string `json:"source_url"`
}

type FieldCheckResult struct {
	Field       string         `json:"field"`
	Expected    interface{}    `json:"expected"`
	Found       interface{}    `json:"found"`
	Status      string         `json:"status"` // Pass, Needs Review, Likely Fail
	Confidence  int            `json:"confidence"`
	Explanation string         `json:"explanation"`
	Rule        RuleBreadcrumb `json:"rule"`
}

type AIEscalation struct {
	Eligible bool   `json:"eligible"`
	Used     bool   `json:"used"`
	Provider string `json:"provider"`
	Reason   string `json:"reason"`
}

type LabelAnalysisResult struct {
	Filename          string                 `json:"filename,omitempty"`
	BeverageType      string                 `json:"beverage_type"`
	OverallStatus     string                 `json:"overall_status"`
	OverallConfidence int                    `json:"overall_confidence"`
	ExtractedFields   ExtractedLabelFields   `json:"extracted_fields"`
	Fields            []FieldCheckResult     `json:"fields"`
	AIEscalation      AIEscalation           `json:"ai_escalation"`
	OCRText           string                 `json:"ocr_text,omitempty"`
	OCR               *ocrclient.OCRResponse `json:"ocr,omitempty"`
	Warnings          []string               `json:"warnings"`
}

type AnalysisInput struct {
	BeverageType   string              `json:"beverage_type"`
	Text           string              `json:"text"`
	ExpectedFields ExpectedLabelFields `json:"expected_fields"`
}
