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
		OverallStatus:  models.StatusMatch,
		Warnings:       []string{},
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
	brandStatus := models.StatusMissingOnLabel
	brandConf := 0
	if validators.FuzzyContains(input.Text, input.ExpectedFields.BrandName) {
		brandFound = input.ExpectedFields.BrandName
		brandStatus = models.StatusMatch
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
	classStatus := models.StatusMissingOnLabel
	classConf := 0
	if validators.FuzzyContains(input.Text, input.ExpectedFields.ClassType) {
		classFound = input.ExpectedFields.ClassType
		classStatus = models.StatusMatch
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
	alcStatus := models.StatusMissingOnLabel
	alcConf := 40
	if validators.FuzzyContains(alcFoundStr, input.ExpectedFields.AlcoholContent) || validators.FuzzyContains(input.ExpectedFields.AlcoholContent, extractedABV) {
		if alcFoundStr != "" {
			alcStatus = models.StatusMatch
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
	netStatus := models.StatusMissingOnLabel
	netConf := 40
	if (validators.FuzzyContains(netContFound, input.ExpectedFields.NetContents) || validators.FuzzyContains(input.ExpectedFields.NetContents, netContFound)) && netContFound != "" {
		netStatus = models.StatusMatch
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

	// 5. Government Warning - strict verbatim check
	govWarningFound, exactCase := validators.ExtractGovernmentWarning(normText)
	govStatus := models.StatusMissingOnLabel
	govConf := 40
	govFoundStr := "Not found"
	govExpStr := "Present and verbatim match required"
	var govDiff *models.GovWarningComparison

	if input.ExpectedFields.GovernmentWarningPresent {
		if govWarningFound {
			fullText := validators.ExtractGovernmentWarningText(normText)
			if fullText != "" {
				comparison := validators.CompareGovernmentWarningVerbatim(fullText)
				govDiff = &comparison
				if comparison.IsExactMatch {
					govStatus = models.StatusMatch
					govFoundStr = "Present (exact match)"
					govConf = 100
				} else if !exactCase {
					govStatus = models.StatusMismatch
					govFoundStr = "Present (incorrect case)"
					govConf = 75
					res.Warnings = append(res.Warnings, "Government warning heading case mismatch. GOVERNMENT WARNING: must be all caps.")
				} else if comparison.Similarity < 0.6 {
					govStatus = models.StatusMismatch
					govFoundStr = "Present (text does not match statute)"
					govConf = 80
					res.Warnings = append(res.Warnings, "Government warning text has very low similarity to statutory requirement.")
				} else {
					govStatus = models.StatusMismatch
					govFoundStr = "Present (text differs from statute)"
					govConf = 85
				}
			} else {
				govStatus = models.StatusUncertain
				govFoundStr = "Heading found, full text unreadable"
				govConf = 60
			}
		}
	} else {
		govExpStr = "Not required"
		if govWarningFound {
			govFoundStr = "Present"
			govStatus = models.StatusUncertain
			govConf = 75
		} else {
			govFoundStr = "Not present"
			govStatus = models.StatusMatch
			govConf = 100
		}
	}

	res.ExtractedFields.GovernmentWarningFound = govWarningFound
	res.Fields = append(res.Fields, models.FieldCheckResult{
		Field:          "Government Warning",
		Expected:       govExpStr,
		Found:          govFoundStr,
		Status:         govStatus,
		Confidence:     govConf,
		Explanation:    "Checked for exact uppercase heading and verbatim statutory text.",
		Rule:           getRule(catalog, "health_warning_statement"),
		GovWarningDiff: govDiff,
	})

	// Overall Status & Confidence
	totalConf := 0
	hasMissingOrMismatch := false
	for _, f := range res.Fields {
		totalConf += f.Confidence
		if f.Status == models.StatusMismatch || f.Status == models.StatusMissingOnLabel {
			res.OverallStatus = models.StatusMismatch
			hasMissingOrMismatch = true
		} else if f.Status == models.StatusUncertain {
			if res.OverallStatus != models.StatusMismatch {
				res.OverallStatus = models.StatusUncertain
			}
			hasMissingOrMismatch = true
		}
	}
	if len(res.Fields) > 0 {
		res.OverallConfidence = totalConf / len(res.Fields)
	}

	if ocrMetadata != nil && ocrMetadata.MeanConfidence > 0 {
		res.OverallConfidence = int(float64(res.OverallConfidence)*0.5 + ocrMetadata.MeanConfidence*0.5)
	}

	// Eligibility logic for escalating OCR-only runs to the AI-native parser.
	isOcrPoor := res.OverallConfidence < 85
	isDenseText := len(input.Text) > 400 && hasMissingOrMismatch
	hasWarningTypo := validators.FuzzyContains(normText, "GOVERMENT") || validators.FuzzyContains(normText, "Surgon") || validators.FuzzyContains(normText, "pregancy")

	if res.OverallStatus != models.StatusMatch {
		res.AISecondRead.Eligible = true
		res.AISecondRead.Reason = "Overall status is not Match."
	} else if hasMissingOrMismatch {
		res.AISecondRead.Eligible = true
		res.AISecondRead.Reason = "Required fields remained missing after local OCR."
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
