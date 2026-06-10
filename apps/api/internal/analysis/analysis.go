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
		BeverageType:  input.BeverageType,
		OCRText:       input.Text,
		OCR:           ocrMetadata,
		OverallStatus: "Pass",
		Warnings:      []string{},
		AIEscalation: models.AIEscalation{
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

	isOcrPoor := false
	if ocrMetadata != nil {
		if ocrMetadata.MeanConfidence > 0 {
			res.OverallConfidence = int(float64(res.OverallConfidence)*0.5 + ocrMetadata.MeanConfidence*0.5)
			if ocrMetadata.MeanConfidence < 65 {
				isOcrPoor = true
			}
		}
	}

	if res.OverallStatus == "Pass" {
		if res.OverallConfidence < 85 && res.OverallConfidence >= 65 {
			res.OverallStatus = "Needs Review"
			hasMissingOrAmbiguous = true
		} else if res.OverallConfidence < 65 {
			res.OverallStatus = "Likely Fail"
			hasMissingOrAmbiguous = true
		}
	}
	
	if ocrMetadata != nil && len(ocrMetadata.Warnings) > 0 {
		res.Warnings = append(res.Warnings, ocrMetadata.Warnings...)
		for _, w := range ocrMetadata.Warnings {
			if w == "Accurate OCR unavailable; fast fallback used. Result requires review." || w == "Deep OCR was not ready. Used fast OCR fallback." || w == "Fast fallback OCR may miss or corrupt label text. Results require review." {
				if res.OverallStatus == "Pass" {
					res.OverallStatus = "Needs Review"
				}
				hasMissingOrAmbiguous = true
			}
		}
	}

	if isOcrPoor {
		res.AIEscalation.Eligible = true
		res.AIEscalation.Reason = "Local OCR providers returned low confidence."
	} else if hasMissingOrAmbiguous {
		res.AIEscalation.Eligible = true
		res.AIEscalation.Reason = "Required fields remained missing after local OCR."
	} else if ocrMetadata != nil && ocrMetadata.Status == "error" {
		res.AIEscalation.Eligible = true
		res.AIEscalation.Reason = "OCR provider was unavailable or errored."
	}

	return res
}
