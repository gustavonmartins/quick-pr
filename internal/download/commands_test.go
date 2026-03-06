package download

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/yourusername/quick-ci/internal/common"
)

func TestBuildCommandReplacesVariables(t *testing.T) {
	// Given: A command template with variables
	template := "git clone --bare {repo} {workdir}/.git"

	// And: Variable values
	vars := CommandVars{
		Repo:     "https://github.com/spacecowboy/Feeder",
		Workdir:  "./workdir",
		PRNumber: 735,
		SHA:      "62ae3465e3d7465d1cb6e07c2125a9f90767467e",
	}

	// When: We build the command
	result := BuildCommand(template, vars)

	// Then: All variables should be replaced
	expected := "git clone --bare https://github.com/spacecowboy/Feeder ./workdir/.git"
	if result != expected {
		t.Errorf("Expected:\n  %s\nGot:\n  %s", expected, result)
	}

	// And: No template variables should remain
	if strings.Contains(result, "{") || strings.Contains(result, "}") {
		t.Errorf("Unreplaced variables in result: %s", result)
	}
}

func TestBuildCommandReplacesAllVariables(t *testing.T) {
	// Given: A command template with all variables
	template := "git -C {workdir} worktree add ./pr-{pr_number} {sha}"

	// And: Variable values
	vars := CommandVars{
		Repo:     "https://github.com/spacecowboy/Feeder",
		Workdir:  "./workdir",
		PRNumber: 735,
		SHA:      "62ae346",
	}

	// When: We build the command
	result := BuildCommand(template, vars)

	// Then: All variables should be replaced
	expected := "git -C ./workdir worktree add ./pr-735 62ae346"
	if result != expected {
		t.Errorf("Expected:\n  %s\nGot:\n  %s", expected, result)
	}
}

func TestBuildSetupCommandReturnsEmptyWhenWorkdirExists(t *testing.T) {
	// Given: A workdir that already exists
	workdir := "../../testdata/existing_workdir"
	if err := os.MkdirAll(workdir+"/.git", 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(workdir) }()

	// And: A setup command template
	template := "git clone --bare {repo} {workdir}/.git"

	vars := CommandVars{
		Repo:    "https://github.com/spacecowboy/Feeder",
		Workdir: workdir,
	}

	// When: We build the setup command
	result := BuildSetupCommand(template, vars)

	// Then: Result should be empty (skip setup)
	if result != "" {
		t.Errorf("Expected empty string when workdir exists, got: %s", result)
	}
}

func TestBuildSetupCommandReturnsCommandWhenWorkdirMissing(t *testing.T) {
	// Given: A workdir that does NOT exist
	workdir := "../../testdata/nonexistent_workdir_for_test"

	// Precondition: directory must not exist
	if _, err := os.Stat(workdir); err == nil {
		t.Fatalf("Test cannot run: directory %s exists. Please remove it manually.", workdir)
	}

	// And: A setup command template
	template := "git clone --bare {repo} {workdir}/.git"

	vars := CommandVars{
		Repo:    "https://github.com/spacecowboy/Feeder",
		Workdir: workdir,
	}

	// When: We build the setup command
	result := BuildSetupCommand(template, vars)

	// Then: Result should be the full command
	expected := "git clone --bare https://github.com/spacecowboy/Feeder ../../testdata/nonexistent_workdir_for_test/.git"
	if result != expected {
		t.Errorf("Expected:\n  %s\nGot:\n  %s", expected, result)
	}
}

func TestPRJsonIncludesParsedCommands(t *testing.T) {
	// Given: A config with git command templates
	config := &Config{
		Repository: "https://github.com/spacecowboy/Feeder",
		Workdir:    "./workdir",
		Setup: []string{
			"git clone --bare {repo} {workdir}/.git",
		},
		PerPR: []string{
			"git -C {workdir} worktree remove ./pr-{pr_number} || true",
			"git -C {workdir} fetch origin pull/{pr_number}/head",
			"git -C {workdir} worktree add ./pr-{pr_number} {sha}",
		},
		Run: []string{
			"go test ./...",
			"go build",
		},
	}

	// And: A PR
	pr := PullRequest{
		Number: 735,
		Title:  "allow searching in feeds",
		Head:   common.Head{Ref: "fix-search-for-feeds", SHA: "62ae346"},
		Base:   common.Base{Ref: "master"},
	}

	// And: A clean output directory
	outputDir := "../../testdata/output_with_commands"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(outputDir) }()

	// When: We save PR with commands
	err := SavePRWithCommands(pr, config, outputDir)
	if err != nil {
		t.Fatalf("Failed to save PR: %v", err)
	}

	// Then: The JSON file should contain parsed commands
	data, err := os.ReadFile(outputDir + "/pr-735.json")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var saved common.PRWithCommands
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify setup command is filled
	expectedSetup := "git clone --bare https://github.com/spacecowboy/Feeder ./workdir/.git"
	if len(saved.Commands.Setup) != 1 || saved.Commands.Setup[0] != expectedSetup {
		t.Errorf("Setup command mismatch.\nExpected: %s\nGot: %v", expectedSetup, saved.Commands.Setup)
	}

	// Verify per_pr commands are filled
	expectedPerPR := []string{
		"git -C ./workdir worktree remove ./pr-735 || true",
		"git -C ./workdir fetch origin pull/735/head",
		"git -C ./workdir worktree add ./pr-735 62ae346",
	}
	if len(saved.Commands.PerPR) != 3 {
		t.Errorf("Expected 3 per_pr commands, got %d", len(saved.Commands.PerPR))
	}
	for i, cmd := range expectedPerPR {
		if saved.Commands.PerPR[i] != cmd {
			t.Errorf("PerPR[%d] mismatch.\nExpected: %s\nGot: %s", i, cmd, saved.Commands.PerPR[i])
		}
	}

	// Verify run commands are included
	if len(saved.Commands.Run) != 2 {
		t.Errorf("Expected 2 run commands, got %d", len(saved.Commands.Run))
	}
}
