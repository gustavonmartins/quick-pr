package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/yourusername/quick-ci/internal/common"
	"github.com/yourusername/quick-ci/internal/run"
)

func TestLoadPRCommandsFromJSON(t *testing.T) {
	prJSON := common.PRWithCommands{
		Number: 735,
		Commands: common.ParsedCommands{
			Setup: []string{"git clone --bare https://example.com/repo ./workdir/.git"},
			PerPR: []string{
				"git -C ./workdir worktree remove ./pr-735 || true",
				"git -C ./workdir fetch origin pull/735/head",
				"git -C ./workdir worktree add ./pr-735 abc123",
			},
			Run: []string{"go test ./...", "go build"},
		},
	}

	tmpDir := "testdata/runner_test"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	data, _ := json.MarshalIndent(prJSON, "", "  ")
	if err := os.WriteFile(tmpDir+"/pr-735.json", data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := run.LoadPRCommands(tmpDir + "/pr-735.json")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if len(loaded.Commands.Setup) != 1 {
		t.Errorf("Expected 1 setup command, got %d", len(loaded.Commands.Setup))
	}
	if len(loaded.Commands.PerPR) != 3 {
		t.Errorf("Expected 3 per_pr commands, got %d", len(loaded.Commands.PerPR))
	}
	if len(loaded.Commands.Run) != 2 {
		t.Errorf("Expected 2 run commands, got %d", len(loaded.Commands.Run))
	}
}

func TestExecuteCommandsWithMock(t *testing.T) {
	var executed []string
	mockExecutor := func(cmd string) error {
		executed = append(executed, cmd)
		return nil
	}

	commands := []string{"cmd1", "cmd2", "cmd3"}

	err := run.ExecuteCommands(commands, mockExecutor)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(executed) != 3 {
		t.Errorf("Expected 3 commands executed, got %d", len(executed))
	}

	for i, cmd := range commands {
		if executed[i] != cmd {
			t.Errorf("Command %d: expected %s, got %s", i, cmd, executed[i])
		}
	}
}

func TestRunCommands_ExecutesMergeBetweenPerPRAndRun(t *testing.T) {
	prJSON := common.PRWithCommands{
		Number: 200,
		Commands: common.ParsedCommands{
			Setup: []string{"setup-cmd"},
			PerPR: []string{"per-pr-cmd"},
			Merge: []string{"merge-cmd"},
			Run:   []string{"run-cmd"},
		},
	}

	tmpDir := "testdata/merge_runner_test"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	data, _ := json.MarshalIndent(prJSON, "", "  ")
	if err := os.WriteFile(tmpDir+"/pr-200.json", data, 0644); err != nil {
		t.Fatal(err)
	}

	var executionOrder []string
	mockExecutor := func(cmd string) error {
		executionOrder = append(executionOrder, cmd)
		return nil
	}

	loaded, err := run.LoadPRCommands(tmpDir + "/pr-200.json")
	if err != nil {
		t.Fatalf("Failed to load PR: %v", err)
	}

	if err := run.ExecuteCommands(loaded.Commands.Setup, mockExecutor); err != nil {
		t.Fatal(err)
	}
	if err := run.ExecuteCommands(loaded.Commands.PerPR, mockExecutor); err != nil {
		t.Fatal(err)
	}
	if err := run.ExecuteCommands(loaded.Commands.Merge, mockExecutor); err != nil {
		t.Fatal(err)
	}
	if err := run.ExecuteCommands(loaded.Commands.Run, mockExecutor); err != nil {
		t.Fatal(err)
	}

	expectedOrder := []string{"setup-cmd", "per-pr-cmd", "merge-cmd", "run-cmd"}

	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d commands executed, got %d: %v", len(expectedOrder), len(executionOrder), executionOrder)
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Command %d: expected '%s', got '%s'", i, expected, executionOrder[i])
		}
	}

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

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
