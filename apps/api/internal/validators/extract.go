package validators

import (
	"regexp"
	"strings"
)

var (
	abvRegex      = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*%\s*(?:Alc\./Vol\.|ALC/VOL|ABV|Alc\.? / Vol\.?)`)
	proofRegex    = regexp.MustCompile(`(?i)\(?(\d+(?:\.\d+)?)\s*Proof\)?`)
	netContRegex  = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(mL|L|fl\s*oz\.?|FL\.?\s*OZ\.?)`)
)

func ExtractABV(text string) string {
	match := abvRegex.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1] + "% Alc./Vol."
	}
	return ""
}

func ExtractProof(text string) string {
	match := proofRegex.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1] + " Proof"
	}
	return ""
}

func ExtractNetContents(text string) string {
	match := netContRegex.FindStringSubmatch(text)
	if len(match) > 0 {
		return match[0]
	}
	return ""
}

func FuzzyContains(text, target string) bool {
	if target == "" {
		return false
	}
	normText := NormalizeForFuzzyMatch(text)
	normTarget := NormalizeForFuzzyMatch(target)
	return strings.Contains(normText, normTarget)
}
