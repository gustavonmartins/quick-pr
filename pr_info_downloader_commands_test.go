package main

import (
	"os"
	"strings"
	"testing"
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
	workdir := "testdata/existing_workdir"
	os.MkdirAll(workdir+"/.git", 0755)
	defer os.RemoveAll(workdir)

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
	workdir := "testdata/nonexistent_workdir_for_test"

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
	expected := "git clone --bare https://github.com/spacecowboy/Feeder testdata/nonexistent_workdir_for_test/.git"
	if result != expected {
		t.Errorf("Expected:\n  %s\nGot:\n  %s", expected, result)
	}
}
