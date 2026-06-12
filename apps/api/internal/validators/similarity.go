package validators

import "strings"

func BigramSimilarity(a, b string) float64 {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}
	if a == b {
		return 1.0
	}

	bigramsA := bigrams(a)
	bigramsB := bigrams(b)

	intersection := 0
	bCopy := make(map[string]int)
	for k, v := range bigramsB {
		bCopy[k] = v
	}
	for bg, count := range bigramsA {
		if bCount, ok := bCopy[bg]; ok {
			min := count
			if bCount < min {
				min = bCount
			}
			intersection += min
		}
	}

	totalA := 0
	for _, v := range bigramsA {
		totalA += v
	}
	totalB := 0
	for _, v := range bigramsB {
		totalB += v
	}

	if totalA+totalB == 0 {
		return 0.0
	}
	return 2.0 * float64(intersection) / float64(totalA+totalB)
}

func bigrams(s string) map[string]int {
	m := make(map[string]int)
	runes := []rune(s)
	for i := 0; i < len(runes)-1; i++ {
		bg := string(runes[i : i+2])
		m[bg]++
	}
	return m
}

func WordOverlap(expected, found string) float64 {
	expWords := strings.Fields(strings.ToLower(expected))
	foundWords := strings.Fields(strings.ToLower(found))
	if len(expWords) == 0 {
		return 0
	}

	foundSet := make(map[string]bool)
	for _, w := range foundWords {
		foundSet[w] = true
	}

	matches := 0
	for _, w := range expWords {
		if foundSet[w] {
			matches++
		}
	}
	return float64(matches) / float64(len(expWords))
}
