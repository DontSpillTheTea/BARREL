package validators

import "testing"

func TestBigramSimilarity(t *testing.T) {
	tests := []struct {
		a, b    string
		minSim  float64
		maxSim  float64
	}{
		{"STONE'S THROW", "Stone's Throw", 0.85, 1.01},
		{"STONE'S THROW", "STONES THROW", 0.80, 1.01},
		{"OLD TOM DISTILLERY", "OLD TOM DISTILLERY", 1.0, 1.01},
		{"OLD TOM DISTILLERY", "old tom distillery", 1.0, 1.01},
		{"completely different", "nothing alike", 0.0, 0.3},
		{"", "", 1.0, 1.01},
		{"something", "", 0.0, 0.01},
	}

	for _, tt := range tests {
		sim := BigramSimilarity(tt.a, tt.b)
		if sim < tt.minSim || sim > tt.maxSim {
			t.Errorf("BigramSimilarity(%q, %q) = %.3f, want [%.2f, %.2f]", tt.a, tt.b, sim, tt.minSim, tt.maxSim)
		}
	}
}

func TestWordOverlap(t *testing.T) {
	tests := []struct {
		expected string
		found    string
		minScore float64
	}{
		{"United States", "Product of United States", 1.0},
		{"United States", "Made in USA", 0.0},
		{"France", "Product of France", 1.0},
		{"United Kingdom", "United Kingdom", 1.0},
	}

	for _, tt := range tests {
		score := WordOverlap(tt.expected, tt.found)
		if score < tt.minScore {
			t.Errorf("WordOverlap(%q, %q) = %.3f, want >= %.2f", tt.expected, tt.found, score, tt.minScore)
		}
	}
}
