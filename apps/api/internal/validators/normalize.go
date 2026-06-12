package validators

import (
	"regexp"
	"strings"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

const CanonicalGovernmentWarning = "GOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink alcoholic beverages during pregnancy because of the risk of birth defects. (2) Consumption of alcoholic beverages impairs your ability to drive a car or operate machinery, and may cause health problems."

func NormalizeWhitespace(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

func NormalizeCase(text string) string {
	return strings.ToLower(text)
}

func NormalizePunctuation(text string) string {
	replacer := strings.NewReplacer(
		"“", "\"", "”", "\"",
		"‘", "'", "’", "'",
		"–", "-", "—", "-",
	)
	return replacer.Replace(text)
}

func NormalizeForFuzzyMatch(text string) string {
	text = NormalizeWhitespace(text)
	text = NormalizeCase(text)
	text = NormalizePunctuation(text)
	re := regexp.MustCompile(`[^a-z0-9]`)
	return re.ReplaceAllString(text, "")
}

func ExtractGovernmentWarning(text string) (found bool, exactCase bool) {
	normText := NormalizeWhitespace(text)
	if strings.Contains(normText, "GOVERNMENT WARNING:") {
		return true, true
	}
	if strings.Contains(strings.ToUpper(normText), "GOVERNMENT WARNING:") {
		return true, false
	}
	return false, false
}

func ExtractGovernmentWarningText(text string) string {
	normText := NormalizeWhitespace(NormalizePunctuation(text))
	idx := strings.Index(normText, "GOVERNMENT WARNING:")
	if idx < 0 {
		upper := strings.ToUpper(normText)
		idx = strings.Index(upper, "GOVERNMENT WARNING:")
		if idx < 0 {
			return ""
		}
	}
	rest := normText[idx:]
	if len(rest) > 350 {
		rest = rest[:350]
	}
	return strings.TrimSpace(rest)
}

func CompareGovernmentWarningVerbatim(extracted string) models.GovWarningComparison {
	canonical := NormalizeWhitespace(CanonicalGovernmentWarning)
	extracted = NormalizeWhitespace(NormalizePunctuation(extracted))

	if extracted == "" {
		return models.GovWarningComparison{
			IsExactMatch:  false,
			ExtractedText: "",
			CanonicalText: canonical,
			Similarity:    0,
		}
	}

	if extracted == canonical {
		return models.GovWarningComparison{
			IsExactMatch:  true,
			ExtractedText: extracted,
			CanonicalText: canonical,
			Similarity:    1.0,
		}
	}

	matches := 0
	shorter := len(canonical)
	if len(extracted) < shorter {
		shorter = len(extracted)
	}
	for i := 0; i < shorter; i++ {
		if canonical[i] == extracted[i] {
			matches++
		}
	}
	longer := len(canonical)
	if len(extracted) > longer {
		longer = len(extracted)
	}
	similarity := 0.0
	if longer > 0 {
		similarity = float64(matches) / float64(longer)
	}

	return models.GovWarningComparison{
		IsExactMatch:  false,
		ExtractedText: extracted,
		CanonicalText: canonical,
		Similarity:    similarity,
	}
}
