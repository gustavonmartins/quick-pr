package run

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/yourusername/quick-ci/internal/common"
)

// TestLoadPRCommandsFromJSON verifies that LoadPRCommands successfully loads
// a PR with commands from a JSON file and parses all command phases correctly.
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

	loaded, err := LoadPRCommands(tmpDir + "/pr-735.json")
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

// TestExecuteCommandsWithMock verifies that ExecuteCommands runs all commands
// in order using a custom executor function and handles the results correctly.
func TestExecuteCommandsWithMock(t *testing.T) {
	var executed []string
	mockExecutor := func(cmd string) error {
		executed = append(executed, cmd)
		return nil
	}

	commands := []string{"cmd1", "cmd2", "cmd3"}

	err := ExecuteCommands(commands, mockExecutor)
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

// TestRunCommands_ExecutesMergeBetweenPerPRAndRun verifies that ExecuteCommands
// executes merge commands in the correct order: after per_pr commands but before run commands.
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

	loaded, err := LoadPRCommands(tmpDir + "/pr-200.json")
	if err != nil {
		t.Fatalf("Failed to load PR: %v", err)
	}

	if err := ExecuteCommands(loaded.Commands.Setup, mockExecutor); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteCommands(loaded.Commands.PerPR, mockExecutor); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteCommands(loaded.Commands.Merge, mockExecutor); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteCommands(loaded.Commands.Run, mockExecutor); err != nil {
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

// indexOf is a test helper that finds the index of an item in a string slice.
func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

// TestExecuteSetupPhase_SkipsWhenWorkdirGitExists verifies that the setup phase
// is skipped when the workdir/.git directory already exists.
func TestExecuteSetupPhase_SkipsWhenWorkdirGitExists(t *testing.T) {
	// Create a temporary workdir with .git directory
	tmpDir := "testdata/skip_setup_test"
	workdir := tmpDir + "/workdir"
	if err := os.MkdirAll(workdir+"/.git", 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	prJSON := common.PRWithCommands{
		Number:  100,
		Workdir: workdir,
		Commands: common.ParsedCommands{
			Setup: []string{"git clone --bare https://example.com/repo " + workdir + "/.git"},
			PerPR: []string{"per-pr-cmd"},
			Run:   []string{"run-cmd"},
		},
	}

	var executed []string
	mockExecutor := func(cmd string) error {
		executed = append(executed, cmd)
		return nil
	}

	// ExecuteSetupPhase should skip setup when workdir/.git exists
	result := ExecuteSetupPhase(prJSON.Commands.Setup, prJSON.Workdir, mockExecutor)

	// Setup commands should NOT have been executed
	for _, cmd := range executed {
		if cmd == prJSON.Commands.Setup[0] {
			t.Errorf("Setup command should have been skipped, but was executed: %s", cmd)
		}
	}

	// The phase result should indicate success (skipped is still success)
	if !result.Success {
		t.Errorf("Expected Success=true for skipped setup, got false")
	}

	// The phase should have no executed commands since it was skipped
	if len(result.Commands) != 0 {
		t.Errorf("Expected 0 commands in skipped setup phase, got %d", len(result.Commands))
	}
}

// TestExecuteSetupPhase_RunsWhenWorkdirGitMissing verifies that the setup phase
// runs when the workdir/.git directory does not exist.
func TestExecuteSetupPhase_RunsWhenWorkdirGitMissing(t *testing.T) {
	// Create a temporary directory without .git
	tmpDir := "testdata/run_setup_test"
	workdir := tmpDir + "/workdir"
	// Ensure the directory does NOT exist
	_ = os.RemoveAll(tmpDir)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	prJSON := common.PRWithCommands{
		Number:  101,
		Workdir: workdir,
		Commands: common.ParsedCommands{
			Setup: []string{"git clone --bare https://example.com/repo " + workdir + "/.git"},
			PerPR: []string{"per-pr-cmd"},
			Run:   []string{"run-cmd"},
		},
	}

	var executed []string
	mockExecutor := func(cmd string) error {
		executed = append(executed, cmd)
		return nil
	}

	// ExecuteSetupPhase should run setup when workdir/.git does not exist
	result := ExecuteSetupPhase(prJSON.Commands.Setup, prJSON.Workdir, mockExecutor)

	// Setup commands SHOULD have been executed
	if len(executed) == 0 {
		t.Error("Expected setup command to be executed, but nothing was executed")
	}

	found := false
	for _, cmd := range executed {
		if cmd == prJSON.Commands.Setup[0] {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected setup command to be executed: %s", prJSON.Commands.Setup[0])
	}

	// The phase result should indicate success
	if !result.Success {
		t.Errorf("Expected Success=true for executed setup, got false")
	}

	// The phase should have the executed commands
	if len(result.Commands) != 1 {
		t.Errorf("Expected 1 command in setup phase, got %d", len(result.Commands))
	}
}
