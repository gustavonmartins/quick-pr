package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CommandVars holds the variables for command template substitution
type CommandVars struct {
	Repo     string
	Workdir  string
	PRNumber int
	SHA      string
}

// BuildCommand replaces template variables with actual values
func BuildCommand(template string, vars CommandVars) string {
	result := template
	result = strings.ReplaceAll(result, "{repo}", vars.Repo)
	result = strings.ReplaceAll(result, "{workdir}", vars.Workdir)
	result = strings.ReplaceAll(result, "{pr_number}", fmt.Sprintf("%d", vars.PRNumber))
	result = strings.ReplaceAll(result, "{sha}", vars.SHA)
	return result
}

// BuildSetupCommand returns the setup command, or empty string if workdir already exists
func BuildSetupCommand(template string, vars CommandVars) string {
	gitDir := vars.Workdir + "/.git"
	if _, err := os.Stat(gitDir); err == nil {
		// workdir/.git exists, skip setup
		return ""
	}
	return BuildCommand(template, vars)
}

// ParsedCommands holds the filled-in command templates
type ParsedCommands struct {
	Setup []string `json:"setup"`
	PerPR []string `json:"per_pr"`
	Run   []string `json:"run"`
}

// PRWithCommands combines PR data with parsed commands
type PRWithCommands struct {
	Number   int            `json:"number"`
	Title    string         `json:"title"`
	State    string         `json:"state"`
	Head     Head           `json:"head"`
	Base     Base           `json:"base"`
	Commits  int            `json:"commits"`
	From     string         `json:"from"`
	To       string         `json:"to"`
	Commands ParsedCommands `json:"commands"`
}

// SavePRWithCommands saves a PR with its parsed commands to a JSON file
func SavePRWithCommands(pr PullRequest, config *Config, outputDir string) error {
	vars := CommandVars{
		Repo:     config.Repository,
		Workdir:  config.Workdir,
		PRNumber: pr.Number,
		SHA:      pr.Head.SHA,
	}

	// Build setup commands
	setupCmds := make([]string, 0, len(config.Setup))
	for _, tmpl := range config.Setup {
		setupCmds = append(setupCmds, BuildCommand(tmpl, vars))
	}

	// Build per_pr commands
	perPRCmds := make([]string, 0, len(config.PerPR))
	for _, tmpl := range config.PerPR {
		perPRCmds = append(perPRCmds, BuildCommand(tmpl, vars))
	}

	// Create PRWithCommands
	prWithCmds := PRWithCommands{
		Number:  pr.Number,
		Title:   pr.Title,
		State:   pr.State,
		Head:    pr.Head,
		Base:    pr.Base,
		Commits: pr.Commits,
		From:    pr.From,
		To:      pr.To,
		Commands: ParsedCommands{
			Setup: setupCmds,
			PerPR: perPRCmds,
			Run:   config.Run,
		},
	}

	// Save to file
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := filepath.Join(outputDir, fmt.Sprintf("pr-%d.json", pr.Number))
	data, err := json.MarshalIndent(prWithCmds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PR: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
