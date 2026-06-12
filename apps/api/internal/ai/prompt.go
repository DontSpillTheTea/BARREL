package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

func BuildPrompt(input SecondReadInput) string {
	expectedJSON, _ := json.MarshalIndent(input.ExpectedFields, "", "  ")

	return fmt.Sprintf(`You are extracting structured fields from a beverage alcohol label image.
Do not decide legal compliance. Do not invent missing values.
Return only JSON matching the schema.
For each field, provide value, confidence, evidence, and reason if missing.
Prefer visible label evidence over assumptions.
If a value is ambiguous, return null or low confidence.
For the government_warning field: transcribe the EXACT text as it appears on the label character by character.
Do not paraphrase, correct, or normalize the warning text. Include all punctuation and capitalization exactly as visible.
The statutory text must begin with "GOVERNMENT WARNING:" in all caps. If any characters differ from what is visible, report exactly what you see.
YOU MUST RETURN ONLY RAW VALID JSON. DO NOT INCLUDE MARKDOWN CODE BLOCKS OR TRIPLE BACKTICKS.

Expected Fields for reference (do not blindly copy them, extract what is visible):
%s

Return exactly this JSON structure:
{
  "candidates": {
    "brand_name": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "image" },
    "class_type": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "image" },
    "alcohol_content": { "abv": "...", "proof": "...", "confidence": 0.99, "evidence": "...", "source": "image" },
    "net_contents": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "image" },
    "producer_or_bottler": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "image" },
    "government_warning": { "present": true, "verbatim_text": "...", "confidence": 0.99, "source": "image", "possible_typos": [] },
    "country_of_origin": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "image" },
    "disclosures": { "value": "...", "confidence": 0.99, "evidence": "...", "source": "image" },
    "image_quality_flags": ["dense_text", "blurry"]
  }
}`, string(expectedJSON))
}

// CleanJSON removes markdown formatting if the model returns it
func CleanJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
