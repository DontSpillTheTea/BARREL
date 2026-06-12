package validators

import "fmt"

// Authorized standards of fill per 27 CFR
var spiritsSizes = map[float64]bool{
	50: true, 100: true, 200: true, 375: true, 500: true,
	750: true, 1000: true, 1750: true,
}

var wineSizes = map[float64]bool{
	100: true, 187: true, 375: true, 500: true, 750: true,
	1000: true, 1500: true, 3000: true,
}

func ValidateContainerSize(netContents, beverageType string) (valid bool, warning string) {
	parsed, ok := parseNetContents(netContents)
	if !ok {
		return true, ""
	}

	ml := toMilliliters(parsed)
	if ml <= 0 {
		return true, ""
	}

	switch beverageType {
	case "distilled_spirits":
		if !spiritsSizes[ml] {
			return false, fmt.Sprintf("%.0f mL is not a standard fill size for distilled spirits per 27 CFR § 5.71.", ml)
		}
	case "wine":
		if !wineSizes[ml] {
			return false, fmt.Sprintf("%.0f mL is not a standard fill size for wine per 27 CFR § 4.72.", ml)
		}
	}
	return true, ""
}
