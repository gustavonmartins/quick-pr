package download

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/yourusername/quick-ci/internal/common"
)

// TestBuildCommandReplacesVariables verifies that BuildCommand correctly replaces
// template variables like {repo} and {workdir} with their actual values.
func TestBuildCommandReplacesVariables(t *testing.T) {
	template := "git clone --bare {repo} {workdir}/.git"
	vars := CommandVars{
		Repo:     "https://github.com/spacecowboy/Feeder",
		Workdir:  "./workdir",
		PRNumber: 735,
		SHA:      "62ae3465e3d7465d1cb6e07c2125a9f90767467e",
	}

	result := BuildCommand(template, vars)

	expected := "git clone --bare https://github.com/spacecowboy/Feeder ./workdir/.git"
	if result != expected {
		t.Errorf("Expected:\n  %s\nGot:\n  %s", expected, result)
	}
	if strings.Contains(result, "{") || strings.Contains(result, "}") {
		t.Errorf("Unreplaced variables in result: %s", result)
	}
}

// TestBuildCommandReplacesAllVariables verifies that BuildCommand can replace
// multiple different variables ({workdir}, {pr_number}, {sha}) in a single template.
func TestBuildCommandReplacesAllVariables(t *testing.T) {
	template := "git -C {workdir} worktree add ./pr-{pr_number} {sha}"
	vars := CommandVars{
		Repo:     "https://github.com/spacecowboy/Feeder",
		Workdir:  "./workdir",
		PRNumber: 735,
		SHA:      "62ae346",
	}

	result := BuildCommand(template, vars)

	expected := "git -C ./workdir worktree add ./pr-735 62ae346"
	if result != expected {
		t.Errorf("Expected:\n  %s\nGot:\n  %s", expected, result)
	}
}

// TestBuildSetupCommandAlwaysReturnsCommand verifies that BuildSetupCommand
// always returns the command with variables replaced, regardless of workdir state.
// The workdir existence check is now handled by the executor (run phase).
func TestBuildSetupCommandAlwaysReturnsCommand(t *testing.T) {
	// Create a workdir with .git to prove the command is returned anyway
	workdir := "../../testdata/setup_always_returns_test"
	if err := os.MkdirAll(workdir+"/.git", 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(workdir) }()

	template := "git clone --bare {repo} {workdir}/.git"
	vars := CommandVars{
		Repo:    "https://github.com/spacecowboy/Feeder",
		Workdir: workdir,
	}

	result := BuildSetupCommand(template, vars)

	// BuildSetupCommand should ALWAYS return the command, even when workdir/.git exists.
	// The workdir existence check should be handled by the executor at runtime.
	expected := "git clone --bare https://github.com/spacecowboy/Feeder " + workdir + "/.git"
	if result != expected {
		t.Errorf("BuildSetupCommand should always return the command.\nExpected:\n  %s\nGot:\n  %s", expected, result)
	}
}

// TestBuildMergeCommands_GeneratesCorrectCommands verifies that BuildMergeCommands
// generates proper git merge commands for the "merge" strategy.
func TestBuildMergeCommands_GeneratesCorrectCommands(t *testing.T) {
	vars := CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 123,
		SHA:      "abc123def456",
	}

	commands := BuildMergeCommands("merge", "main", vars)

	if len(commands) == 0 {
		t.Fatal("Expected merge commands, got empty slice")
	}

	found := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "git") && strings.Contains(cmd, "merge") && strings.Contains(cmd, "main") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected a git merge command with base branch 'main', got: %v", commands)
	}
}

// TestBuildMergeCommands_NoneStrategy_ReturnsEmpty verifies that BuildMergeCommands
// returns an empty slice when the merge strategy is "none".
func TestBuildMergeCommands_NoneStrategy_ReturnsEmpty(t *testing.T) {
	vars := CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 123,
		SHA:      "abc123def456",
	}

	commands := BuildMergeCommands("none", "main", vars)

	if len(commands) != 0 {
		t.Errorf("Expected empty commands for 'none' strategy, got: %v", commands)
	}
}

// TestBuildMergeCommands_RebaseStrategy verifies that BuildMergeCommands
// generates proper git rebase commands for the "rebase" strategy.
func TestBuildMergeCommands_RebaseStrategy(t *testing.T) {
	vars := CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 456,
		SHA:      "def789ghi012",
	}

	commands := BuildMergeCommands("rebase", "develop", vars)

	if len(commands) == 0 {
		t.Fatal("Expected rebase commands, got empty slice")
	}

	found := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "git") && strings.Contains(cmd, "rebase") && strings.Contains(cmd, "develop") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected a git rebase command with base branch 'develop', got: %v", commands)
	}
}

// TestBuildMergeCommands_SquashStrategy verifies that BuildMergeCommands
// generates proper git merge --squash commands for the "squash" strategy.
func TestBuildMergeCommands_SquashStrategy(t *testing.T) {
	vars := CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 789,
		SHA:      "ghi345jkl678",
	}

	commands := BuildMergeCommands("squash", "master", vars)

	if len(commands) == 0 {
		t.Fatal("Expected squash commands, got empty slice")
	}

	found := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "git") && strings.Contains(cmd, "merge") && strings.Contains(cmd, "--squash") && strings.Contains(cmd, "master") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected a git merge --squash command with base branch 'master', got: %v", commands)
	}
}

// TestPRJsonIncludesParsedCommands verifies that SavePRWithCommands properly parses
// command templates and includes the parsed commands (setup, per_pr, run) in the saved JSON.
func TestPRJsonIncludesParsedCommands(t *testing.T) {
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

	pr := PullRequest{
		Number: 735,
		Title:  "allow searching in feeds",
		Head:   common.Head{Ref: "fix-search-for-feeds", SHA: "62ae346"},
		Base:   common.Base{Ref: "master"},
	}

	outputDir := "../../testdata/output_with_commands"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(outputDir) }()

	err := SavePRWithCommands(pr, config, outputDir)
	if err != nil {
		t.Fatalf("Failed to save PR: %v", err)
	}

	data, err := os.ReadFile(outputDir + "/pr-735.json")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var saved common.PRWithCommands
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	expectedSetup := "git clone --bare https://github.com/spacecowboy/Feeder ./workdir/.git"
	if len(saved.Commands.Setup) != 1 || saved.Commands.Setup[0] != expectedSetup {
		t.Errorf("Setup command mismatch.\nExpected: %s\nGot: %v", expectedSetup, saved.Commands.Setup)
	}

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

	if len(saved.Commands.Run) != 2 {
		t.Errorf("Expected 2 run commands, got %d", len(saved.Commands.Run))
	}
}

// TestSavePRWithCommands_IncludesMergeCommands verifies that SavePRWithCommands
// includes merge commands in the saved JSON when a merge strategy is configured.
func TestSavePRWithCommands_IncludesMergeCommands(t *testing.T) {
	pr := PullRequest{
		Number:  100,
		Title:   "Test PR",
		State:   "open",
		From:    "feature-branch",
		To:      "main",
		Commits: 3,
		Head: common.Head{
			Ref: "feature-branch",
			SHA: "abc123",
		},
		Base: common.Base{
			Ref: "main",
		},
	}

	config := &Config{
		Repository:    "https://github.com/example/repo",
		Workdir:       "./workdir",
		MergeStrategy: "merge",
		Run:           []string{"go test ./..."},
	}

	outputDir := "../../testdata/merge_output"
	_ = os.RemoveAll(outputDir)
	defer func() { _ = os.RemoveAll(outputDir) }()

	err := SavePRWithCommands(pr, config, outputDir)
	if err != nil {
		t.Fatalf("Failed to save PR: %v", err)
	}

	data, err := os.ReadFile(outputDir + "/pr-100.json")
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var savedPR common.PRWithCommands
	if err := json.Unmarshal(data, &savedPR); err != nil {
		t.Fatalf("Failed to parse saved JSON: %v", err)
	}

	if len(savedPR.Commands.Merge) == 0 {
		t.Error("Expected merge commands in saved PR, got empty slice")
	}
}
