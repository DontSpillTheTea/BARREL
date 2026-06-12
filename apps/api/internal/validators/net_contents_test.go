package validators

import (
	"testing"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

func TestCompareNetContents(t *testing.T) {
	tests := []struct {
		name       string
		expected   string
		found      string
		wantStatus string
	}{
		{"exact match", "750 mL", "750 mL", models.StatusMatch},
		{"case difference", "750 mL", "750 ml", models.StatusMatch},
		{"no space", "750 mL", "750ml", models.StatusMatch},
		{"different values", "750 mL", "1 L", models.StatusMismatch},
		{"liters vs ml", "1 L", "1000 mL", models.StatusMatch},
		{"missing on label", "750 mL", "", models.StatusMissingOnLabel},
		{"missing in app", "", "750 mL", models.StatusMissingInApp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, _ := CompareNetContents(tt.expected, tt.found)
			if status != tt.wantStatus {
				t.Errorf("CompareNetContents(%q, %q) = %q, want %q", tt.expected, tt.found, status, tt.wantStatus)
			}
		})
	}
}
