package main

import (
	"encoding/json"
	"os"
	"testing"
)

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
		Head:   Head{Ref: "fix-search-for-feeds", SHA: "62ae346"},
		Base:   Base{Ref: "master"},
	}

	// And: A clean output directory
	outputDir := "testdata/output_with_commands"
	os.MkdirAll(outputDir, 0755)
	defer os.RemoveAll(outputDir)

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

	var saved PRWithCommands
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