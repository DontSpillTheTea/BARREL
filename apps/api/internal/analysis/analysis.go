package analysis

import (
	"fmt"
	"math"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/ocr/providers"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/rules"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/validators"
)

func abs(v float64) float64 {
	return math.Abs(v)
}

func getRule(catalog *rules.Catalog, id string) models.RuleBreadcrumb {
	if r, ok := catalog.RulesByID[id]; ok {
		var citation string
		if r.CFR.Section != nil && r.CFR.Section != "" {
			citation = fmt.Sprintf("27 CFR Part %v Section %v", r.CFR.Part, r.CFR.Section)
		} else {
			citation = fmt.Sprintf("27 CFR Part %v", r.CFR.Part)
		}
		return models.RuleBreadcrumb{ID: r.ID, Citation: citation, SourceURL: r.SourceURL}
	}
	return models.RuleBreadcrumb{}
}

func AnalyzeText(input models.AnalysisInput, catalog *rules.Catalog, ocrMetadata *providers.OCRResult) models.LabelAnalysisResult {
	res := models.LabelAnalysisResult{
		BeverageType:   input.BeverageType,
		OCRText:        input.Text,
		OCR:            ocrMetadata,
		ExpectedFields: input.ExpectedFields,
		OverallStatus:  "Pass",
		Warnings:      []string{},
		AISecondRead: &models.AISecondRead{
			Eligible: false,
			Used:     false,
			Provider: "none",
			Reason:   "",
		},
	}

	normText := validators.NormalizeWhitespace(input.Text)

	prefix := ""
	if input.BeverageType == "distilled_spirits" {
		prefix = "spirits_"
	} else if input.BeverageType == "wine" {
		prefix = "wine_"
	} else if input.BeverageType == "malt_beverages" {
		prefix = "malt_"
	}

	// 1. Brand Name
	brandFound := ""
	brandStatus := "Likely Fail"
	brandConf := 0
	if validators.FuzzyContains(input.Text, input.ExpectedFields.BrandName) {
		brandFound = input.ExpectedFields.BrandName
		brandStatus = "Pass"
		brandConf = 95
	} else if input.ExpectedFields.BrandName != "" {
		brandFound = "Not found"
		brandConf = 40
	}

	res.ExtractedFields.BrandName = brandFound
	res.Fields = append(res.Fields, models.FieldCheckResult{
		Field:       "Brand Name",
		Expected:    input.ExpectedFields.BrandName,
		Found:       brandFound,
		Status:      brandStatus,
		Confidence:  brandConf,
		Explanation: "Checked via fuzzy normalized substring match.",
		Rule:        getRule(catalog, prefix+"brand_name"),
	})

	// 2. Class/Type
	classFound := ""
	classStatus := "Likely Fail"
	classConf := 0
	if validators.FuzzyContains(input.Text, input.ExpectedFields.ClassType) {
		classFound = input.ExpectedFields.ClassType
		classStatus = "Pass"
		classConf = 95
	} else if input.ExpectedFields.ClassType != "" {
		classFound = "Not found"
		classConf = 40
	}
	res.ExtractedFields.ClassType = classFound
	res.Fields = append(res.Fields, models.FieldCheckResult{
		Field:       "Class/Type",
		Expected:    input.ExpectedFields.ClassType,
		Found:       classFound,
		Status:      classStatus,
		Confidence:  classConf,
		Explanation: "Checked via fuzzy normalized substring match.",
		Rule:        getRule(catalog, prefix+"class_type"),
	})

	// 3. Alcohol Content
	extractedABV := validators.ExtractABV(input.Text)
	extractedProof := validators.ExtractProof(input.Text)
	alcFoundStr := extractedABV
	if extractedProof != "" {
		if alcFoundStr != "" {
			alcFoundStr += " (" + extractedProof + ")"
		} else {
			alcFoundStr = extractedProof
		}
	}
	alcStatus := "Likely Fail"
	alcConf := 40
	if validators.FuzzyContains(alcFoundStr, input.ExpectedFields.AlcoholContent) || validators.FuzzyContains(input.ExpectedFields.AlcoholContent, extractedABV) {
		if alcFoundStr != "" {
			alcStatus = "Pass"
			alcConf = 95
		}
	}
	if alcFoundStr == "" {
		alcFoundStr = "Not found"
	}

	res.ExtractedFields.AlcoholContent = alcFoundStr
	res.Fields = append(res.Fields, models.FieldCheckResult{
		Field:       "Alcohol Content",
		Expected:    input.ExpectedFields.AlcoholContent,
		Found:       alcFoundStr,
		Status:      alcStatus,
		Confidence:  alcConf,
		Explanation: "Extracted using regex patterns for ABV and Proof.",
		Rule:        getRule(catalog, prefix+"alcohol_content"),
	})

	// 4. Net Contents
	netContFound := validators.ExtractNetContents(input.Text)
	netStatus := "Likely Fail"
	netConf := 40
	if (validators.FuzzyContains(netContFound, input.ExpectedFields.NetContents) || validators.FuzzyContains(input.ExpectedFields.NetContents, netContFound)) && netContFound != "" {
		netStatus = "Pass"
		netConf = 95
	}
	if netContFound == "" {
		netContFound = "Not found"
	}
	res.ExtractedFields.NetContents = netContFound
	res.Fields = append(res.Fields, models.FieldCheckResult{
		Field:       "Net Contents",
		Expected:    input.ExpectedFields.NetContents,
		Found:       netContFound,
		Status:      netStatus,
		Confidence:  netConf,
		Explanation: "Extracted using regex patterns for volume.",
		Rule:        getRule(catalog, prefix+"net_contents"),
	})

	// 5. Government Warning
	govWarningFound, exactCase := validators.ExtractGovernmentWarning(normText)
	govStatus := "Likely Fail"
	govConf := 95
	govFoundStr := "Not found"
	govExpStr := "Present and exact case"

	if input.ExpectedFields.GovernmentWarningPresent {
		if govWarningFound {
			if exactCase {
				govStatus = "Pass"
				govFoundStr = "Present (Exact Case)"
				govConf = 100
			} else {
				govStatus = "Needs Review"
				govFoundStr = "Present (Incorrect Case)"
				res.Warnings = append(res.Warnings, "Government warning heading case mismatch.")
				govConf = 75
			}
		} else {
			govConf = 40
		}
	} else {
		govExpStr = "Not required"
		if govWarningFound {
			govFoundStr = "Present"
			govStatus = "Needs Review"
			govConf = 75
		} else {
			govFoundStr = "Not present"
			govStatus = "Pass"
			govConf = 100
		}
	}

	res.ExtractedFields.GovernmentWarningFound = govWarningFound
	res.Fields = append(res.Fields, models.FieldCheckResult{
		Field:       "Government Warning",
		Expected:    govExpStr,
		Found:       govFoundStr,
		Status:      govStatus,
		Confidence:  govConf,
		Explanation: "Checked for exact uppercase heading and presence.",
		Rule:        getRule(catalog, "health_warning_statement"),
	})

	// Overall Status & Confidence
	totalConf := 0
	hasMissingOrAmbiguous := false
	for _, f := range res.Fields {
		totalConf += f.Confidence
		if f.Status == "Likely Fail" {
			res.OverallStatus = "Likely Fail"
			hasMissingOrAmbiguous = true
		} else if f.Status == "Needs Review" {
			if res.OverallStatus != "Likely Fail" {
				res.OverallStatus = "Needs Review"
			}
			hasMissingOrAmbiguous = true
		}
	}
	if len(res.Fields) > 0 {
		res.OverallConfidence = totalConf / len(res.Fields)
	}

	if ocrMetadata != nil && ocrMetadata.MeanConfidence > 0 {
		res.OverallConfidence = int(float64(res.OverallConfidence)*0.5 + ocrMetadata.MeanConfidence*0.5)
	}

	// Eligibility logic for AI Second Read
	isOcrPoor := false
	if res.OverallConfidence < 85 {
		isOcrPoor = true
	}

	// dense text / high OCR text volume but missing key fields
	isDenseText := len(input.Text) > 400 && hasMissingOrAmbiguous

	// Warning detected but not exact case
	warningNotExact := res.ExtractedFields.GovernmentWarningFound && govStatus == "Needs Review"

	// Typo candidate
	hasWarningTypo := validators.FuzzyContains(normText, "GOVERMENT") || validators.FuzzyContains(normText, "Surgon") || validators.FuzzyContains(normText, "pregancy")

	if res.OverallStatus != "Pass" {
		res.AISecondRead.Eligible = true
		res.AISecondRead.Reason = "Overall status is not Pass."
	} else if hasMissingOrAmbiguous {
		res.AISecondRead.Eligible = true
		res.AISecondRead.Reason = "Required fields remained missing after local OCR."
	} else if warningNotExact {
		res.AISecondRead.Eligible = true
		res.AISecondRead.Reason = "Government warning detected but case is not exact."
	} else if hasWarningTypo {
		res.AISecondRead.Eligible = true
		res.AISecondRead.Reason = "Suspicious typos detected in Government Warning."
	} else if isDenseText {
		res.AISecondRead.Eligible = true
		res.AISecondRead.Reason = "High text density with missing or ambiguous fields."
	} else if isOcrPoor {
		res.AISecondRead.Eligible = true
		res.AISecondRead.Reason = "Azure Vision OCR returned low confidence."
	} else if ocrMetadata != nil && ocrMetadata.Status == "error" {
		res.AISecondRead.Eligible = true
		res.AISecondRead.Reason = "Azure Vision OCR was unavailable or errored."
	}

	return res
}
