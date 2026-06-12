package validators

import "testing"

func TestValidateContainerSize(t *testing.T) {
	tests := []struct {
		name         string
		netContents  string
		beverageType string
		wantValid    bool
	}{
		{"750mL spirits valid", "750 mL", "distilled_spirits", true},
		{"1000mL spirits valid", "1 L", "distilled_spirits", true},
		{"375mL spirits valid", "375 mL", "distilled_spirits", true},
		{"600mL spirits invalid", "600 mL", "distilled_spirits", false},
		{"750mL wine valid", "750 mL", "wine", true},
		{"500mL wine valid", "500 mL", "wine", true},
		{"600mL wine invalid", "600 mL", "wine", false},
		{"any size malt valid", "355 mL", "malt_beverages", true},
		{"empty string", "", "distilled_spirits", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := ValidateContainerSize(tt.netContents, tt.beverageType)
			if valid != tt.wantValid {
				t.Errorf("ValidateContainerSize(%q, %q) valid=%v, want %v", tt.netContents, tt.beverageType, valid, tt.wantValid)
			}
		})
	}
}
