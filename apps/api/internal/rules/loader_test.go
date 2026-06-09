package rules

import (
	"testing"
)

func TestLoadCatalog(t *testing.T) {
	catalog, err := LoadCatalog("../../../../rules/ttb")
	if err != nil {
		t.Fatalf("Failed to load rules: %v", err)
	}
	if len(catalog.RulesByID) == 0 {
		t.Error("Expected to load rules, got 0")
	}
	if _, ok := catalog.RulesByID["health_warning_statement"]; !ok {
		t.Error("Expected health_warning_statement rule")
	}
}
