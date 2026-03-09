package download

import (
	"os"
	"testing"
)

// TestConfigWithMergeStrategy_LoadsCorrectly verifies that LoadConfig properly
// parses the merge_strategy field from a JSON configuration file.
func TestConfigWithMergeStrategy_LoadsCorrectly(t *testing.T) {
	configJSON := `{
		"repository": "https://github.com/example/repo",
		"workdir": "./workdir",
		"merge_strategy": "merge",
		"run": ["go test ./..."]
	}`

	tmpFile := "../../testdata/config_with_merge_strategy.json"
	if err := os.MkdirAll("testdata", 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	err := os.WriteFile(tmpFile, []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.MergeStrategy != "merge" {
		t.Errorf("Expected merge_strategy 'merge', got '%s'", config.MergeStrategy)
	}
}
