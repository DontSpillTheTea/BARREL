package rules

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type CFR struct {
	Title   int         `yaml:"title"`
	Part    interface{} `yaml:"part"`
	Section interface{} `yaml:"section"`
}

type Rule struct {
	ID             string   `yaml:"id"`
	Requirement    string   `yaml:"requirement"`
	BeverageTypes  []string `yaml:"beverage_types"`
	CFR            CFR      `yaml:"cfr"`
	SourceURL      string   `yaml:"source_url"`
	CheckType      string   `yaml:"check_type"`
	Severity       string   `yaml:"severity"`
	Explanation    string   `yaml:"explanation"`
	RequiredText   string   `yaml:"required_text,omitempty"`
}

type Ruleset struct {
	RulesetID      string `yaml:"ruleset_id"`
	RulesetVersion string `yaml:"ruleset_version"`
	LastReviewed   string `yaml:"last_reviewed"`
	Disclaimer     string `yaml:"disclaimer"`
	Rules          []Rule `yaml:"rules"`
}

type Catalog struct {
	RulesByID map[string]Rule
}

func LoadCatalog(rulesDir string) (*Catalog, error) {
	catalog := &Catalog{
		RulesByID: make(map[string]Rule),
	}

	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules directory: %w", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".yaml" || filepath.Ext(entry.Name()) == ".yml" {
			data, err := os.ReadFile(filepath.Join(rulesDir, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read rule file %s: %w", entry.Name(), err)
			}

			var ruleset Ruleset
			if err := yaml.Unmarshal(data, &ruleset); err != nil {
				return nil, fmt.Errorf("failed to parse rule file %s: %w", entry.Name(), err)
			}

			for _, rule := range ruleset.Rules {
				catalog.RulesByID[rule.ID] = rule
			}
		}
	}

	return catalog, nil
}
