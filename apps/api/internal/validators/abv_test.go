package validators

import (
	"testing"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

func TestCompareABV_WithinTolerance(t *testing.T) {
	tests := []struct {
		name         string
		expected     string
		found        string
		beverageType string
		wantStatus   string
	}{
		{"exact match spirits", "45%", "45% Alc./Vol.", "distilled_spirits", models.StatusMatch},
		{"within 0.3% spirits", "45%", "44.8% Alc./Vol.", "distilled_spirits", models.StatusMatch},
		{"outside 0.3% spirits", "45%", "43%", "distilled_spirits", models.StatusMismatch},
		{"within 1.0% wine high", "15%", "14.5%", "wine", models.StatusMatch},
		{"within 1.5% wine low", "12%", "10.8%", "wine", models.StatusMatch},
		{"outside tolerance wine", "15%", "12%", "wine", models.StatusMismatch},
		{"within 0.3% malt", "5.5%", "5.3%", "malt_beverages", models.StatusMatch},
		{"missing on label", "45%", "", "distilled_spirits", models.StatusMissingOnLabel},
		{"missing in app", "", "45%", "distilled_spirits", models.StatusMissingInApp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, _ := CompareABV(tt.expected, tt.found, tt.beverageType)
			if status != tt.wantStatus {
				t.Errorf("CompareABV(%q, %q, %q) = %q, want %q", tt.expected, tt.found, tt.beverageType, status, tt.wantStatus)
			}
		})
	}
}

func TestParseABVNumeric(t *testing.T) {
	tests := []struct {
		input string
		want  float64
		ok    bool
	}{
		{"45% Alc./Vol.", 45.0, true},
		{"44.8%", 44.8, true},
		{"12.5% ABV", 12.5, true},
		{"no number", 0, false},
		{"", 0, false},
	}
	for _, tt := range tests {
		val, ok := ParseABVNumeric(tt.input)
		if ok != tt.ok || (ok && val != tt.want) {
			t.Errorf("ParseABVNumeric(%q) = (%v, %v), want (%v, %v)", tt.input, val, ok, tt.want, tt.ok)
		}
	}
}
