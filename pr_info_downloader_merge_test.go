package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestConfigWithMergeStrategy_LoadsCorrectly(t *testing.T) {
	// Given: A config file with merge_strategy field
	configJSON := `{
		"repository": "https://github.com/example/repo",
		"workdir": "./workdir",
		"merge_strategy": "merge",
		"run": ["go test ./..."]
	}`

	tmpFile := "testdata/config_with_merge_strategy.json"
	os.MkdirAll("testdata", 0755)
	defer os.Remove(tmpFile)

	err := os.WriteFile(tmpFile, []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// When: We load the config
	config, err := LoadConfig(tmpFile)

	// Then: The config should load successfully with the merge strategy
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.MergeStrategy != "merge" {
		t.Errorf("Expected merge_strategy 'merge', got '%s'", config.MergeStrategy)
	}
}

func TestBuildMergeCommands_GeneratesCorrectCommands(t *testing.T) {
	// Given: PR data with a base branch
	vars := CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 123,
		SHA:      "abc123def456",
	}

	baseBranch := "main"
	strategy := "merge"

	// When: We build merge commands
	commands := BuildMergeCommands(strategy, baseBranch, vars)

	// Then: Commands should include git merge with the base branch
	if len(commands) == 0 {
		t.Fatal("Expected merge commands, got empty slice")
	}

	// Verify the merge command references the base branch
	found := false
	for _, cmd := range commands {
		if containsAll(cmd, "git", "merge", "main") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected a git merge command with base branch 'main', got: %v", commands)
	}
}

func TestBuildMergeCommands_NoneStrategy_ReturnsEmpty(t *testing.T) {
	// Given: Strategy is "none"
	vars := CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 123,
		SHA:      "abc123def456",
	}

	baseBranch := "main"
	strategy := "none"

	// When: We build merge commands
	commands := BuildMergeCommands(strategy, baseBranch, vars)

	// Then: Commands should be empty
	if len(commands) != 0 {
		t.Errorf("Expected empty commands for 'none' strategy, got: %v", commands)
	}
}

func TestBuildMergeCommands_RebaseStrategy(t *testing.T) {
	// Given: Strategy is "rebase"
	vars := CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 456,
		SHA:      "def789ghi012",
	}

	baseBranch := "develop"
	strategy := "rebase"

	// When: We build merge commands
	commands := BuildMergeCommands(strategy, baseBranch, vars)

	// Then: Commands should include git rebase with the base branch
	if len(commands) == 0 {
		t.Fatal("Expected rebase commands, got empty slice")
	}

	// Verify the rebase command references the base branch
	found := false
	for _, cmd := range commands {
		if containsAll(cmd, "git", "rebase", "develop") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected a git rebase command with base branch 'develop', got: %v", commands)
	}
}

func TestBuildMergeCommands_SquashStrategy(t *testing.T) {
	// Given: Strategy is "squash"
	vars := CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 789,
		SHA:      "ghi345jkl678",
	}

	baseBranch := "master"
	strategy := "squash"

	// When: We build merge commands
	commands := BuildMergeCommands(strategy, baseBranch, vars)

	// Then: Commands should include git merge --squash with the base branch
	if len(commands) == 0 {
		t.Fatal("Expected squash commands, got empty slice")
	}

	// Verify the squash merge command references the base branch
	found := false
	for _, cmd := range commands {
		if containsAll(cmd, "git", "merge", "--squash", "master") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected a git merge --squash command with base branch 'master', got: %v", commands)
	}
}

func TestSavePRWithCommands_IncludesMergeCommands(t *testing.T) {
	// Given: A PR and config with merge_strategy
	pr := PullRequest{
		Number:  100,
		Title:   "Test PR",
		State:   "open",
		From:    "feature-branch",
		To:      "main",
		Commits: 3,
		Head: Head{
			Ref: "feature-branch",
			SHA: "abc123",
		},
		Base: Base{
			Ref: "main",
		},
	}

	config := &Config{
		Repository:    "https://github.com/example/repo",
		Workdir:       "./workdir",
		MergeStrategy: "merge",
		Run:           []string{"go test ./..."},
	}

	// And: A clean output directory
	outputDir := "testdata/merge_output"
	os.RemoveAll(outputDir)
	defer os.RemoveAll(outputDir)

	// When: We save the PR with commands
	err := SavePRWithCommands(pr, config, outputDir)

	// Then: The saved JSON should include merge commands
	if err != nil {
		t.Fatalf("Failed to save PR: %v", err)
	}

	// Load the saved file
	data, err := os.ReadFile(outputDir + "/pr-100.json")
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var savedPR PRWithCommands
	if err := json.Unmarshal(data, &savedPR); err != nil {
		t.Fatalf("Failed to parse saved JSON: %v", err)
	}

	// Verify merge commands are present
	if len(savedPR.Commands.Merge) == 0 {
		t.Error("Expected merge commands in saved PR, got empty slice")
	}
}

func TestRunCommands_ExecutesMergeBetweenPerPRAndRun(t *testing.T) {
	// Given: A PR with setup, per_pr, merge, and run commands
	prJSON := PRWithCommands{
		Number: 200,
		Commands: ParsedCommands{
			Setup: []string{"setup-cmd"},
			PerPR: []string{"per-pr-cmd"},
			Merge: []string{"merge-cmd"},
			Run:   []string{"run-cmd"},
		},
	}

	// Write to temp file
	tmpDir := "testdata/merge_runner_test"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	data, _ := json.MarshalIndent(prJSON, "", "  ")
	os.WriteFile(tmpDir+"/pr-200.json", data, 0644)

	// And: A mock executor that records execution order
	var executionOrder []string
	mockExecutor := func(cmd string) error {
		executionOrder = append(executionOrder, cmd)
		return nil
	}

	// When: We load and execute commands
	loaded, err := LoadPRCommands(tmpDir + "/pr-200.json")
	if err != nil {
		t.Fatalf("Failed to load PR: %v", err)
	}

	// Execute in order: setup, per_pr, merge, run
	ExecuteCommands(loaded.Commands.Setup, mockExecutor)
	ExecuteCommands(loaded.Commands.PerPR, mockExecutor)
	ExecuteCommands(loaded.Commands.Merge, mockExecutor)
	ExecuteCommands(loaded.Commands.Run, mockExecutor)

	// Then: Commands should be executed in the correct order
	expectedOrder := []string{"setup-cmd", "per-pr-cmd", "merge-cmd", "run-cmd"}

	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d commands executed, got %d: %v", len(expectedOrder), len(executionOrder), executionOrder)
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Command %d: expected '%s', got '%s'", i, expected, executionOrder[i])
		}
	}

	// Verify merge comes after per_pr and before run
	perPRIndex := indexOf(executionOrder, "per-pr-cmd")
	mergeIndex := indexOf(executionOrder, "merge-cmd")
	runIndex := indexOf(executionOrder, "run-cmd")

	if mergeIndex <= perPRIndex {
		t.Errorf("Merge should execute after per_pr: per_pr at %d, merge at %d", perPRIndex, mergeIndex)
	}

	if mergeIndex >= runIndex {
		t.Errorf("Merge should execute before run: merge at %d, run at %d", mergeIndex, runIndex)
	}
}

// Helper function to check if a string contains all substrings
func containsAll(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if !containsString(s, sub) {
			return false
		}
	}
	return true
}

// Helper function to check if a string contains a substring
func containsString(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStringImpl(s, sub))
}

func containsStringImpl(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Helper function to find index of string in slice
func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
