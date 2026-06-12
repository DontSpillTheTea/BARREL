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
CRITICAL for government_warning:
- Do NOT reconstruct the warning text from memory. Transcribe ONLY what is visible.
- If the small body text is not fully legible, set body_verbatim to empty and body_confidence to 0.
- Do NOT fill in the canonical statutory text if you cannot read it clearly.
- Set legibility to "clear", "partial", or "illegible" based on how readable the warning text is.
- prefix_seen: whether you see any form of "GOVERNMENT WARNING" heading.
- prefix_exact_caps: whether the prefix is exactly "GOVERNMENT WARNING:" in all caps.
- prefix_bold: whether the prefix text appears in bold/heavy weight styling (required per 27 CFR § 16.22).
- body_verbatim: the exact text after the prefix, character by character as visible. Leave empty if illegible.
- verbatim_text: the full warning including prefix, exactly as visible.
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
    "government_warning": { "present": true, "prefix_seen": true, "prefix_exact_caps": true, "prefix_bold": true, "verbatim_text": "...", "body_verbatim": "...", "body_confidence": 0.95, "legibility": "clear", "confidence": 0.99, "source": "image", "possible_typos": [] },
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
