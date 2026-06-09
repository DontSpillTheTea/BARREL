package validators

import (
	"regexp"
	"strings"
)

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
