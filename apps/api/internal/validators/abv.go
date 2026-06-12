package validators

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

var abvNumericRegex = regexp.MustCompile(`(\d+(?:\.\d+)?)`)

func ParseABVNumeric(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	match := abvNumericRegex.FindStringSubmatch(s)
	if len(match) < 2 {
		return 0, false
	}
	val, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, false
	}
	return val, true
}

func ABVTolerance(beverageType string, abv float64) float64 {
	switch beverageType {
	case "wine":
		if abv > 14.0 {
			return 1.0 // 27 CFR § 4.36
		}
		return 1.5 // 27 CFR § 4.36
	case "distilled_spirits":
		return 0.3 // 27 CFR § 5.37
	case "malt_beverages":
		return 0.3 // 27 CFR § 7.71
	default:
		return 0.3
	}
}

func CompareABV(expected, found, beverageType string) (status string, similarity float64, explanation string) {
	if found == "" || found == "Not found" {
		if expected != "" {
			return models.StatusMissingOnLabel, 0, "Alcohol content not found on label."
		}
		return models.StatusMissingOnLabel, 0, ""
	}
	if expected == "" {
		return models.StatusMissingInApp, 0.9, "Alcohol content found on label but no expected value provided."
	}

	expVal, expOk := ParseABVNumeric(expected)
	foundVal, foundOk := ParseABVNumeric(found)

	if !expOk || !foundOk {
		sim := BigramSimilarity(expected, found)
		if sim >= 0.85 {
			return models.StatusMatch, sim, "Alcohol content matches (text comparison)."
		}
		return models.StatusMismatch, sim, "Alcohol content could not be parsed numerically; text comparison failed."
	}

	diff := math.Abs(expVal - foundVal)
	tolerance := ABVTolerance(beverageType, expVal)

	if diff <= tolerance {
		return models.StatusMatch, 1.0 - (diff / (tolerance * 2)), "Alcohol content within regulatory tolerance."
	}

	return models.StatusMismatch, 1.0 - (diff / 100.0), "Alcohol content exceeds regulatory tolerance."
}
