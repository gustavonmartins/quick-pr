package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadPRCommandsFromJSON(t *testing.T) {
	// Given: A PR JSON file with commands
	prJSON := PRWithCommands{
		Number: 735,
		Commands: ParsedCommands{
			Setup: []string{"git clone --bare https://example.com/repo ./workdir/.git"},
			PerPR: []string{
				"git -C ./workdir worktree remove ./pr-735 || true",
				"git -C ./workdir fetch origin pull/735/head",
				"git -C ./workdir worktree add ./pr-735 abc123",
			},
			Run: []string{"go test ./...", "go build"},
		},
	}

	// Write to temp file
	tmpDir := "testdata/runner_test"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	data, _ := json.MarshalIndent(prJSON, "", "  ")
	os.WriteFile(tmpDir+"/pr-735.json", data, 0644)

	// When: We load commands from the file
	loaded, err := LoadPRCommands(tmpDir + "/pr-735.json")

	// Then: Commands are parsed correctly
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
	// Given: A mock executor that records calls
	var executed []string
	mockExecutor := func(cmd string) error {
		executed = append(executed, cmd)
		return nil
	}

	commands := []string{"cmd1", "cmd2", "cmd3"}

	// When: We execute commands
	err := ExecuteCommands(commands, mockExecutor)

	// Then: All commands were executed in order
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
