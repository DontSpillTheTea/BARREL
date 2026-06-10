package models

import "github.com/DontSpillTheTea/barrel/apps/api/internal/ocr/providers"

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

type AISecondRead struct {
	Eligible   bool                   `json:"eligible"`
	Used       bool                   `json:"used"`
	Provider   string                 `json:"provider"`
	Reason     string                 `json:"reason,omitempty"`
	Status     string                 `json:"status,omitempty"`
	RawText    string                 `json:"raw_text,omitempty"`
	Candidates AISecondReadCandidates `json:"candidates,omitempty"`
	Findings   []AISecondReadFinding  `json:"findings,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

type AISecondReadCandidates struct {
	BrandName             string `json:"brand_name,omitempty"`
	ClassType             string `json:"class_type,omitempty"`
	AlcoholContent        string `json:"alcohol_content,omitempty"`
	NetContents           string `json:"net_contents,omitempty"`
	GovernmentWarningText string `json:"government_warning_text,omitempty"`
	BottlerNameAddress    string `json:"bottler_name_address,omitempty"`
	CountryOfOrigin       string `json:"country_of_origin,omitempty"`
	ColorFlavorDisclosure string `json:"color_flavor_disclosure,omitempty"`
}

type AISecondReadFinding struct {
	Field    string `json:"field"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Evidence string `json:"evidence,omitempty"`
}

type LabelAnalysisResult struct {
	Filename          string                 `json:"filename,omitempty"`
	RequestedProvider string                 `json:"requested_provider,omitempty"`
	BeverageType      string                 `json:"beverage_type"`
	OverallStatus     string                 `json:"overall_status"`
	OverallConfidence int                    `json:"overall_confidence"`
	ExpectedFields    ExpectedLabelFields    `json:"expected_fields"`
	ExtractedFields   ExtractedLabelFields   `json:"extracted_fields"`
	Fields            []FieldCheckResult     `json:"fields"`
	AIEscalation      AIEscalation           `json:"ai_escalation"`
	AISecondRead      *AISecondRead          `json:"ai_second_read,omitempty"`
	OCRText           string                 `json:"ocr_text,omitempty"`
	OCR               *providers.OCRResult   `json:"ocr,omitempty"`
	Warnings          []string               `json:"warnings"`
}

type AnalysisInput struct {
	BeverageType   string              `json:"beverage_type"`
	Text           string              `json:"text"`
	ExpectedFields ExpectedLabelFields `json:"expected_fields"`
}

type ReviewSummary struct {
	ID                string `json:"id"`
	JobID             string `json:"job_id"`
	Filename          string `json:"filename"`
	SubmittedAt       string `json:"submitted_at"`
	CompletedAt       string `json:"completed_at,omitempty"`
	OCRProvider       string `json:"ocr_provider"`
	OverallStatus     string `json:"overall_status"`
	OverallConfidence int    `json:"overall_confidence"`
	FieldPassCount    int    `json:"field_pass_count"`
	FieldTotalCount   int    `json:"field_total_count"`
	ReviewerDecision  string `json:"reviewer_decision,omitempty"`
	BeverageType      string `json:"beverage_type"`
	IsBatch           bool   `json:"is_batch"`
	BatchID           string `json:"batch_id,omitempty"`
}

type ReviewDetail struct {
	Summary          ReviewSummary       `json:"summary"`
	Result           LabelAnalysisResult `json:"result"`
	OriginalImageURL string              `json:"original_image_url,omitempty"`
	RawOCRText       string              `json:"raw_ocr_text"`
}
