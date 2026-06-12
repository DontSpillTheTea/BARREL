package validators

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

var netContentsNumericRegex = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(ml|l|fl\.?\s*oz\.?|oz\.?)`)

type parsedVolume struct {
	Value float64
	Unit  string
}

func parseNetContents(s string) (parsedVolume, bool) {
	s = strings.TrimSpace(s)
	match := netContentsNumericRegex.FindStringSubmatch(s)
	if len(match) < 3 {
		return parsedVolume{}, false
	}
	val, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return parsedVolume{}, false
	}
	unit := strings.ToLower(strings.TrimSpace(match[2]))
	unit = strings.ReplaceAll(unit, " ", "")
	unit = strings.ReplaceAll(unit, ".", "")
	return parsedVolume{Value: val, Unit: unit}, true
}

func toMilliliters(v parsedVolume) float64 {
	switch {
	case v.Unit == "ml":
		return v.Value
	case v.Unit == "l":
		return v.Value * 1000
	case strings.Contains(v.Unit, "oz"):
		return v.Value * 29.5735
	default:
		return v.Value
	}
}

func CompareNetContents(expected, found string) (status string, similarity float64, explanation string) {
	if found == "" || found == "Not found" {
		if expected != "" {
			return models.StatusMissingOnLabel, 0, "Net contents not found on label."
		}
		return models.StatusMissingOnLabel, 0, ""
	}
	if expected == "" {
		return models.StatusMissingInApp, 0.9, "Net contents found on label but no expected value provided."
	}

	expParsed, expOk := parseNetContents(expected)
	foundParsed, foundOk := parseNetContents(found)

	if !expOk || !foundOk {
		sim := BigramSimilarity(expected, found)
		if sim >= 0.85 {
			return models.StatusMatch, sim, "Net contents match (text comparison)."
		}
		return models.StatusMismatch, sim, "Net contents could not be parsed; text comparison used."
	}

	expMl := toMilliliters(expParsed)
	foundMl := toMilliliters(foundParsed)

	if expMl == 0 {
		return models.StatusUncertain, 0.5, "Expected net contents value is zero."
	}

	diff := math.Abs(expMl - foundMl)
	tolerance := expMl * 0.01 // 1% tolerance for unit conversion rounding

	if diff <= tolerance {
		return models.StatusMatch, 1.0, "Net contents match."
	}

	return models.StatusMismatch, 1.0 - (diff / expMl), "Net contents do not match."
}
