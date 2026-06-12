package models

import "github.com/DontSpillTheTea/barrel/apps/api/internal/ocr/providers"

const (
	StatusMatch          = "Match"
	StatusMismatch       = "Mismatch"
	StatusMissingOnLabel = "Missing on Label"
	StatusMissingInApp   = "Missing in Application Data"
	StatusUncertain      = "Uncertain"
)

type ExpectedLabelFields struct {
	BrandName                string `json:"brand_name,omitempty"`
	ClassType                string `json:"class_type,omitempty"`
	AlcoholContent           string `json:"alcohol_content,omitempty"`
	NetContents              string `json:"net_contents,omitempty"`
	GovernmentWarningPresent bool   `json:"government_warning_present,omitempty"`
	ProducerBottler          string `json:"producer_bottler,omitempty"`
	CountryOfOrigin          string `json:"country_of_origin,omitempty"`
}

type ExtractedLabelFields struct {
	BrandName              string `json:"brand_name,omitempty"`
	ClassType              string `json:"class_type,omitempty"`
	AlcoholContent         string `json:"alcohol_content,omitempty"`
	NetContents            string `json:"net_contents,omitempty"`
	GovernmentWarningFound bool   `json:"government_warning_found,omitempty"`
	ProducerBottler        string `json:"producer_bottler,omitempty"`
	CountryOfOrigin        string `json:"country_of_origin,omitempty"`
}

type GovWarningComparison struct {
	IsExactMatch  bool    `json:"is_exact_match"`
	ExtractedText string  `json:"extracted_text"`
	CanonicalText string  `json:"canonical_text"`
	Similarity    float64 `json:"similarity"`
}

type RuleBreadcrumb struct {
	ID        string `json:"id"`
	Citation  string `json:"citation"`
	SourceURL string `json:"source_url"`
}

type FieldCheckResult struct {
	Field          string                `json:"field"`
	Expected       interface{}           `json:"expected"`
	Found          interface{}           `json:"found"`
	Status         string                `json:"status"`
	Confidence     int                   `json:"confidence"`
	Similarity     float64               `json:"similarity,omitempty"`
	AIConfidence   float64               `json:"ai_confidence,omitempty"`
	Explanation    string                `json:"explanation"`
	Rule           RuleBreadcrumb        `json:"rule"`
	GovWarningDiff *GovWarningComparison `json:"gov_warning_diff,omitempty"`
}

type AIEscalation struct {
	Eligible bool   `json:"eligible"`
	Used     bool   `json:"used"`
	Provider string `json:"provider"`
	Reason   string `json:"reason"`
}

type AIExtractedField struct {
	Value           string  `json:"value"`
	Confidence      float64 `json:"confidence"`
	Evidence        string  `json:"evidence,omitempty"`
	ReasonIfMissing string  `json:"reason_if_missing,omitempty"`
	Source          string  `json:"source"` // typically "image"
}

type AIGovernmentWarning struct {
	Present         bool    `json:"present"`
	VerbatimText    string  `json:"verbatim_text,omitempty"`
	Confidence      float64 `json:"confidence"`
	PossibleTypos   []string`json:"possible_typos,omitempty"`
	Source          string  `json:"source"`
	ReasonIfMissing string  `json:"reason_if_missing,omitempty"`
}

type AIAlcoholContent struct {
	ABV             string  `json:"abv"`
	Proof           string  `json:"proof,omitempty"`
	Confidence      float64 `json:"confidence"`
	Evidence        string  `json:"evidence,omitempty"`
	Source          string  `json:"source"`
	ReasonIfMissing string  `json:"reason_if_missing,omitempty"`
}

type AINativeExtraction struct {
	BrandName          AIExtractedField    `json:"brand_name"`
	ClassType          AIExtractedField    `json:"class_type"`
	AlcoholContent     AIAlcoholContent    `json:"alcohol_content"`
	NetContents        AIExtractedField    `json:"net_contents"`
	ProducerOrBottler  AIExtractedField    `json:"producer_or_bottler"`
	GovernmentWarning  AIGovernmentWarning `json:"government_warning"`
	CountryOfOrigin    AIExtractedField    `json:"country_of_origin"`
	Disclosures        AIExtractedField    `json:"disclosures"`
	ImageQualityFlags  []string            `json:"image_quality_flags,omitempty"`
}

type AISecondRead struct {
	Eligible   bool                `json:"eligible"`
	Used       bool                `json:"used"`
	Provider   string              `json:"provider"`
	Reason     string              `json:"reason,omitempty"`
	Status     string              `json:"status,omitempty"`
	RawText    string              `json:"raw_text,omitempty"`
	Candidates AINativeExtraction  `json:"candidates,omitempty"`
	Error      string              `json:"error,omitempty"`
}

type LabelAnalysisResult struct {
	Filename          string                 `json:"filename,omitempty"`
	RequestedProvider string                 `json:"requested_provider,omitempty"`
	BeverageType      string                 `json:"beverage_type"`
	OverallStatus     string                 `json:"overall_status"`
	OverallConfidence int                    `json:"overall_confidence"`
	ProcessingTimeMs  int64                  `json:"processing_time_ms,omitempty"`
	ExpectedFields    ExpectedLabelFields    `json:"expected_fields"`
	ExtractedFields   ExtractedLabelFields   `json:"extracted_fields"`
	Fields            []FieldCheckResult     `json:"fields"`
	AIEscalation      AIEscalation           `json:"ai_escalation"`
	AISecondRead      *AISecondRead          `json:"ai_second_read,omitempty"`
	OCRText           string                 `json:"ocr_text,omitempty"`
	OCR               *providers.OCRResult   `json:"ocr,omitempty"`
	ImageQualityFlags []string               `json:"image_quality_flags,omitempty"`
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
	ProviderRequested string `json:"provider_requested,omitempty"`
	ProviderUsed      string `json:"provider_used"`
	OverallStatus     string `json:"overall_status"`
	OverallConfidence int    `json:"overall_confidence"`
	FieldPassCount    int    `json:"field_pass_count"`
	FieldTotalCount   int    `json:"field_total_count"`
	ReviewerDecision  string `json:"reviewer_decision,omitempty"`
	BeverageType      string `json:"beverage_type"`
	IsBatch           bool   `json:"is_batch"`
	BatchID           string `json:"batch_id,omitempty"`
	BrandName         string `json:"brand_name,omitempty"`
	ClassType         string `json:"class_type,omitempty"`
	AlcoholContent    string `json:"alcohol_content,omitempty"`
	NetContents       string `json:"net_contents,omitempty"`
}

type ReviewDetail struct {
	Summary          ReviewSummary       `json:"summary"`
	Result           LabelAnalysisResult `json:"result"`
	OriginalImageURL string              `json:"original_image_url,omitempty"`
	RawOCRText       string              `json:"raw_ocr_text"`
}
