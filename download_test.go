package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/yourusername/quick-ci/internal/common"
	"github.com/yourusername/quick-ci/internal/download"
)

// === commands_test.go tests ===

func TestBuildCommandReplacesVariables(t *testing.T) {
	template := "git clone --bare {repo} {workdir}/.git"
	vars := download.CommandVars{
		Repo:     "https://github.com/spacecowboy/Feeder",
		Workdir:  "./workdir",
		PRNumber: 735,
		SHA:      "62ae3465e3d7465d1cb6e07c2125a9f90767467e",
	}

	result := download.BuildCommand(template, vars)

	expected := "git clone --bare https://github.com/spacecowboy/Feeder ./workdir/.git"
	if result != expected {
		t.Errorf("Expected:\n  %s\nGot:\n  %s", expected, result)
	}
	if strings.Contains(result, "{") || strings.Contains(result, "}") {
		t.Errorf("Unreplaced variables in result: %s", result)
	}
}

func TestBuildCommandReplacesAllVariables(t *testing.T) {
	template := "git -C {workdir} worktree add ./pr-{pr_number} {sha}"
	vars := download.CommandVars{
		Repo:     "https://github.com/spacecowboy/Feeder",
		Workdir:  "./workdir",
		PRNumber: 735,
		SHA:      "62ae346",
	}

	result := download.BuildCommand(template, vars)

	expected := "git -C ./workdir worktree add ./pr-735 62ae346"
	if result != expected {
		t.Errorf("Expected:\n  %s\nGot:\n  %s", expected, result)
	}
}

func TestBuildSetupCommandReturnsEmptyWhenWorkdirExists(t *testing.T) {
	workdir := "testdata/existing_workdir"
	if err := os.MkdirAll(workdir+"/.git", 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(workdir) }()

	template := "git clone --bare {repo} {workdir}/.git"
	vars := download.CommandVars{
		Repo:    "https://github.com/spacecowboy/Feeder",
		Workdir: workdir,
	}

	result := download.BuildSetupCommand(template, vars)

	if result != "" {
		t.Errorf("Expected empty string when workdir exists, got: %s", result)
	}
}

func TestBuildSetupCommandReturnsCommandWhenWorkdirMissing(t *testing.T) {
	workdir := "testdata/nonexistent_workdir_for_test"

	if _, err := os.Stat(workdir); err == nil {
		t.Fatalf("Test cannot run: directory %s exists. Please remove it manually.", workdir)
	}

	template := "git clone --bare {repo} {workdir}/.git"
	vars := download.CommandVars{
		Repo:    "https://github.com/spacecowboy/Feeder",
		Workdir: workdir,
	}

	result := download.BuildSetupCommand(template, vars)

	expected := "git clone --bare https://github.com/spacecowboy/Feeder testdata/nonexistent_workdir_for_test/.git"
	if result != expected {
		t.Errorf("Expected:\n  %s\nGot:\n  %s", expected, result)
	}
}

func TestPRJsonIncludesParsedCommands(t *testing.T) {
	config := &download.Config{
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

	pr := download.PullRequest{
		Number: 735,
		Title:  "allow searching in feeds",
		Head:   common.Head{Ref: "fix-search-for-feeds", SHA: "62ae346"},
		Base:   common.Base{Ref: "master"},
	}

	outputDir := "testdata/output_with_commands"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(outputDir) }()

	err := download.SavePRWithCommands(pr, config, outputDir)
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

// === github_test.go tests ===

func TestFetchOpenPRsFromFeederRepo(t *testing.T) {
	config, err := download.LoadConfig("testdata/feeder_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	prs, err := download.FetchPullRequests(config)
	if err != nil {
		t.Fatalf("Failed to fetch PRs: %v", err)
	}

	if len(prs) != 19 {
		t.Errorf("Expected 19 open PRs, got %d", len(prs))
	}
}

func TestFetchPR1030Details(t *testing.T) {
	config, err := download.LoadConfig("testdata/feeder_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expected, err := loadExpectedPR("testdata/pr_1030_expected.json")
	if err != nil {
		t.Fatalf("Failed to load expected output: %v", err)
	}

	prs, err := download.FetchPullRequests(config)
	if err != nil {
		t.Fatalf("Failed to fetch PRs: %v", err)
	}

	var pr1030 *download.PullRequest
	for i := range prs {
		if prs[i].Number == 1030 {
			pr1030 = &prs[i]
			break
		}
	}

	if pr1030 == nil {
		t.Fatal("PR #1030 not found")
	}

	if pr1030.Number != expected.Number {
		t.Errorf("Number: expected %d, got %d", expected.Number, pr1030.Number)
	}
	if pr1030.From != expected.From {
		t.Errorf("From: expected '%s', got '%s'", expected.From, pr1030.From)
	}
	if pr1030.To != expected.To {
		t.Errorf("To: expected '%s', got '%s'", expected.To, pr1030.To)
	}
	if pr1030.Commits != expected.Commits {
		t.Errorf("Commits: expected %d, got %d", expected.Commits, pr1030.Commits)
	}
}

func loadExpectedPR(filepath string) (*download.PullRequest, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var pr download.PullRequest
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

func TestSavePRsToOutputFolder(t *testing.T) {
	config, err := download.LoadConfig("testdata/feeder_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	outputDir := "testdata/output"
	_ = os.RemoveAll(outputDir)

	prs, err := download.FetchPullRequests(config)
	if err != nil {
		t.Fatalf("Failed to fetch PRs: %v", err)
	}

	err = download.SavePRsToFiles(prs, outputDir)
	if err != nil {
		t.Fatalf("Failed to save PRs: %v", err)
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(files) != 19 {
		t.Errorf("Expected 19 files, got %d", len(files))
	}

	for _, pr := range prs {
		filename := fmt.Sprintf("%s/pr-%d.json", outputDir, pr.Number)
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Errorf("Failed to read file for PR #%d: %v", pr.Number, err)
			continue
		}

		var savedPR download.PullRequest
		if err := json.Unmarshal(data, &savedPR); err != nil {
			t.Errorf("Failed to parse JSON for PR #%d: %v", pr.Number, err)
			continue
		}

		if savedPR.Number != pr.Number {
			t.Errorf("PR #%d: Number mismatch, got %d", pr.Number, savedPR.Number)
		}
		if savedPR.From != pr.From {
			t.Errorf("PR #%d: From mismatch, expected '%s', got '%s'", pr.Number, pr.From, savedPR.From)
		}
		if savedPR.To != pr.To {
			t.Errorf("PR #%d: To mismatch, expected '%s', got '%s'", pr.Number, pr.To, savedPR.To)
		}
		if savedPR.Commits != pr.Commits {
			t.Errorf("PR #%d: Commits mismatch, expected %d, got %d", pr.Number, pr.Commits, savedPR.Commits)
		}
	}
}

// === merge_test.go tests ===

func TestConfigWithMergeStrategy_LoadsCorrectly(t *testing.T) {
	configJSON := `{
		"repository": "https://github.com/example/repo",
		"workdir": "./workdir",
		"merge_strategy": "merge",
		"run": ["go test ./..."]
	}`

	tmpFile := "testdata/config_with_merge_strategy.json"
	if err := os.MkdirAll("testdata", 0755); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	err := os.WriteFile(tmpFile, []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := download.LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.MergeStrategy != "merge" {
		t.Errorf("Expected merge_strategy 'merge', got '%s'", config.MergeStrategy)
	}
}

func TestBuildMergeCommands_GeneratesCorrectCommands(t *testing.T) {
	vars := download.CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 123,
		SHA:      "abc123def456",
	}

	commands := download.BuildMergeCommands("merge", "main", vars)

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

func TestBuildMergeCommands_NoneStrategy_ReturnsEmpty(t *testing.T) {
	vars := download.CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 123,
		SHA:      "abc123def456",
	}

	commands := download.BuildMergeCommands("none", "main", vars)

	if len(commands) != 0 {
		t.Errorf("Expected empty commands for 'none' strategy, got: %v", commands)
	}
}

func TestBuildMergeCommands_RebaseStrategy(t *testing.T) {
	vars := download.CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 456,
		SHA:      "def789ghi012",
	}

	commands := download.BuildMergeCommands("rebase", "develop", vars)

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

func TestBuildMergeCommands_SquashStrategy(t *testing.T) {
	vars := download.CommandVars{
		Repo:     "https://github.com/example/repo",
		Workdir:  "./workdir",
		PRNumber: 789,
		SHA:      "ghi345jkl678",
	}

	commands := download.BuildMergeCommands("squash", "master", vars)

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

func TestSavePRWithCommands_IncludesMergeCommands(t *testing.T) {
	pr := download.PullRequest{
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

	config := &download.Config{
		Repository:    "https://github.com/example/repo",
		Workdir:       "./workdir",
		MergeStrategy: "merge",
		Run:           []string{"go test ./..."},
	}

	outputDir := "testdata/merge_output"
	_ = os.RemoveAll(outputDir)
	defer func() { _ = os.RemoveAll(outputDir) }()

	err := download.SavePRWithCommands(pr, config, outputDir)
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
