package download

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/quick-ci/internal/common"
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

// BuildMergeCommands generates git commands based on the merge strategy
func BuildMergeCommands(strategy, baseBranch string, vars CommandVars) []string {
	worktree := fmt.Sprintf("%s/pr-%d", vars.Workdir, vars.PRNumber)
	fetchCmd := fmt.Sprintf("git -C %s fetch origin %s:refs/remotes/origin/%s",
		vars.Workdir, baseBranch, baseBranch)
	switch strategy {
	case "none", "":
		return []string{}
	case "merge":
		return []string{
			fetchCmd,
			fmt.Sprintf("git -C %s merge --no-edit origin/%s", worktree, baseBranch),
		}
	case "rebase":
		return []string{
			fetchCmd,
			fmt.Sprintf("git -C %s rebase origin/%s", worktree, baseBranch),
		}
	case "squash":
		return []string{
			fetchCmd,
			fmt.Sprintf("git -C %s merge --squash --no-edit origin/%s", worktree, baseBranch),
		}
	default:
		return []string{}
	}
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

	// Build merge commands
	mergeCmds := BuildMergeCommands(config.MergeStrategy, pr.Base.Ref, vars)

	// Create PRWithCommands
	prWithCmds := common.PRWithCommands{
		Number:  pr.Number,
		Title:   pr.Title,
		State:   pr.State,
		Head:    pr.Head,
		Base:    pr.Base,
		Commits: pr.Commits,
		From:    pr.From,
		To:      pr.To,
		Commands: common.ParsedCommands{
			Setup: setupCmds,
			PerPR: perPRCmds,
			Merge: mergeCmds,
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
