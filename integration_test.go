package main

import (
	"os"
	"testing"

	"github.com/yourusername/quick-ci/internal/common"
	"github.com/yourusername/quick-ci/internal/download"
	"github.com/yourusername/quick-ci/internal/run"
)

// Integration tests that verify packages work together

func TestDownloadPackage_LoadConfig(t *testing.T) {
	config, err := download.LoadConfig("testdata/feeder_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.Repository == "" {
		t.Error("Expected repository to be set")
	}
}

func TestDownloadPackage_BuildCommand(t *testing.T) {
	vars := download.CommandVars{
		Repo:     "https://github.com/test/repo",
		Workdir:  "./workdir",
		PRNumber: 123,
		SHA:      "abc123",
	}
	result := download.BuildCommand("git clone {repo} {workdir}", vars)
	expected := "git clone https://github.com/test/repo ./workdir"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestDownloadPackage_BuildMergeCommands(t *testing.T) {
	vars := download.CommandVars{
		Workdir:  "./workdir",
		PRNumber: 100,
	}
	cmds := download.BuildMergeCommands("merge", "main", vars)
	if len(cmds) == 0 {
		t.Error("Expected merge commands")
	}
}

func TestRunPackage_ExecuteCommands(t *testing.T) {
	var executed []string
	mockExecutor := func(cmd string) error {
		executed = append(executed, cmd)
		return nil
	}

	err := run.ExecuteCommands([]string{"cmd1", "cmd2"}, mockExecutor)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(executed) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(executed))
	}
}

func TestRunPackage_ExecutePhase(t *testing.T) {
	result := run.ExecutePhase("test", []string{"echo hello"})
	if !result.Success {
		t.Errorf("Expected success, got failure: %v", result)
	}
	if len(result.Commands) != 1 {
		t.Errorf("Expected 1 command result, got %d", len(result.Commands))
	}
}

func TestRunPackage_LoadPRCommands(t *testing.T) {
	// Create temp file
	tmpDir := "testdata/integration_test"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	content := `{"number": 1, "title": "Test", "commands": {"setup": ["echo setup"], "run": ["echo run"]}}`
	if err := os.WriteFile(tmpDir+"/pr-1.json", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	pr, err := run.LoadPRCommands(tmpDir + "/pr-1.json")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}
	if pr.Number != 1 {
		t.Errorf("Expected PR number 1, got %d", pr.Number)
	}
}

func TestCommonTypes_PRResult(t *testing.T) {
	result := common.PRResult{
		Number:  100,
		Title:   "Test PR",
		Success: true,
		Phases: []common.PhaseResult{
			{Name: "setup", Success: true},
			{Name: "run", Success: true},
		},
	}
	if !result.Success {
		t.Error("Expected success")
	}
	if len(result.Phases) != 2 {
		t.Errorf("Expected 2 phases, got %d", len(result.Phases))
	}
}
