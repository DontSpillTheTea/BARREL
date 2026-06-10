package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

func BuildPrompt(input SecondReadInput) string {
	expectedJSON, _ := json.MarshalIndent(input.ExpectedFields, "", "  ")

	return fmt.Sprintf(`You are BARREL, an expert TTB (Alcohol and Tobacco Tax and Trade Bureau) label reviewer assistant.
Your task is to analyze the provided label image and return structured candidate evidence comparing what is visible to the expected fields.

CRITICAL INSTRUCTIONS:
1. Extract ONLY what is visible on the label. Do not invent or hallucinate missing text.
2. Preserve exact spelling and punctuation, especially for the Government Warning statement.
3. Compare the visible text to the expected fields.
4. Identify likely TTB review concerns and add them to the "findings" array. Use severity "Warning" or "Error".
5. Provide a short "evidence" snippet for each finding if applicable.
6. Make it clear this is advisory evidence, not a final legal determination.
7. YOU MUST RETURN ONLY RAW VALID JSON. DO NOT INCLUDE MARKDOWN CODE BLOCKS OR TRIPLE BACKTICKS.

Expected Fields:
%s

Return exactly this JSON structure (and nothing else):
{
	"candidates": {
		"brand_name": "...",
		"class_type": "...",
		"alcohol_content": "...",
		"net_contents": "...",
		"government_warning_text": "...",
		"bottler_name_address": "...",
		"country_of_origin": "...",
		"color_flavor_disclosure": "..."
	},
	"findings": [
		{
			"field": "Government Warning",
			"severity": "Warning",
			"message": "Missing 'GOVERNMENT WARNING:' heading.",
			"evidence": "..."
		}
	]
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
