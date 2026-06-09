package validators

import "testing"

func TestExtractABV(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"45% Alc./Vol.", "45% Alc./Vol."},
		{"13.5% ABV", "13.5% Alc./Vol."},
		{"No alcohol here", ""},
	}
	for _, tc := range tests {
		res := ExtractABV(tc.input)
		if res != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, res)
		}
	}
}

func TestExtractProof(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"(90 Proof)", "90 Proof"},
		{"100 PROOF", "100 Proof"},
		{"No proof here", ""},
	}
	for _, tc := range tests {
		res := ExtractProof(tc.input)
		if res != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, res)
		}
	}
}

func TestExtractNetContents(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"750 mL", "750 mL"},
		{"12 FL. OZ.", "12 FL. OZ."},
		{"1 L", "1 L"},
	}
	for _, tc := range tests {
		res := ExtractNetContents(tc.input)
		if res != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, res)
		}
	}
}

func TestExtractGovernmentWarning(t *testing.T) {
	found, exact := ExtractGovernmentWarning("GOVERNMENT WARNING: (1) According to...")
	if !found || !exact {
		t.Error("Expected exact match")
	}

	found, exact = ExtractGovernmentWarning("Government Warning: (1) According to...")
	if !found || exact {
		t.Error("Expected found but not exact case")
	}
}

func TestFuzzyContains(t *testing.T) {
	if !FuzzyContains("OLD TOM DISTILLERY", "Old Tom Distillery") {
		t.Error("Expected case-insensitive match")
	}
	if FuzzyContains("40% ABV", "45% ABV") {
		t.Error("Expected mismatch")
	}
}
