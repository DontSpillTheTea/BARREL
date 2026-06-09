package analysis

import (
	"testing"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/rules"
)

func TestAnalyzeText(t *testing.T) {
	catalog, _ := rules.LoadCatalog("../../../../rules/ttb")

	input := models.AnalysisInput{
		BeverageType: "distilled_spirits",
		Text:         "OLD TOM DISTILLERY Kentucky Straight Bourbon Whiskey 45% Alc./Vol. (90 Proof) 750 mL GOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink alcoholic beverages during pregnancy because of the risk of birth defects. (2) Consumption of alcoholic beverages impairs your ability to drive a car or operate machinery, and may cause health problems.",
		ExpectedFields: models.ExpectedLabelFields{
			BrandName:                "OLD TOM DISTILLERY",
			ClassType:                "Kentucky Straight Bourbon Whiskey",
			AlcoholContent:           "45% Alc./Vol. (90 Proof)",
			NetContents:              "750 mL",
			GovernmentWarningPresent: true,
		},
	}

	res := AnalyzeText(input, catalog, nil)

	if res.OverallStatus != "Pass" {
		t.Errorf("Expected Pass, got %s", res.OverallStatus)
	}
	if res.OverallConfidence < 85 {
		t.Errorf("Expected confidence >= 85, got %d", res.OverallConfidence)
	}

	foundBrand := false
	for _, f := range res.Fields {
		if f.Field == "Brand Name" {
			foundBrand = true
			if f.Status != "Pass" {
				t.Errorf("Expected Brand Name to pass, got %s", f.Status)
			}
		}
	}
	if !foundBrand {
		t.Error("Expected Brand Name field check")
	}
}
