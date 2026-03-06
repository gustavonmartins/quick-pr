package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/quick-ci/internal/download"
	"github.com/yourusername/quick-ci/internal/run"
)

func main() {
	configPath := flag.String("config", "", "path to config file (download mode)")
	outputDir := flag.String("output", "./output", "directory for PR JSON files")
	runDir := flag.String("run", "", "directory with PR JSON files to execute (run mode)")
	flag.Parse()

	// Run mode: execute commands from JSON files
	if *runDir != "" {
		runCommands(*runDir)
		return
	}

	// Download mode: fetch PRs and save JSON files
	if *configPath != "" {
		downloadPRs(*configPath, *outputDir)
		return
	}

	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  Download PRs: quick-ci -config <config.json> -output <dir>\n")
	fmt.Fprintf(os.Stderr, "  Run commands: quick-ci -run <dir>\n")
	os.Exit(1)
}

func downloadPRs(configPath, outputDir string) {
	config, err := download.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetching PRs from %s...\n", config.Repository)
	prs, err := download.FetchPullRequests(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching PRs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d open PRs\n", len(prs))

	for _, pr := range prs {
		err := download.SavePRWithCommands(pr, config, outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving PR #%d: %v\n", pr.Number, err)
			os.Exit(1)
		}
		fmt.Printf("Saved pr-%d.json\n", pr.Number)
	}

	fmt.Printf("Done. Saved %d PR files to %s\n", len(prs), outputDir)
}

func runCommands(dir string) {
	files, err := filepath.Glob(filepath.Join(dir, "pr-*.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No pr-*.json files found in %s\n", dir)
		os.Exit(1)
	}

	fmt.Printf("Found %d PR files\n", len(files))

	for _, file := range files {
		processPRFile(file)
	}

	fmt.Printf("\nDone. Processed %d PRs\n", len(files))
}

func processPRFile(file string) {
	pr, err := run.LoadPRCommands(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", file, err)
		os.Exit(1)
	}

	fmt.Printf("\n=== PR #%d: %s ===\n", pr.Number, pr.Title)

	executePhase("setup", pr.Commands.Setup)
	executePhase("per-PR", pr.Commands.PerPR)
	executePhase("merge", pr.Commands.Merge)
	executePhase("CI", pr.Commands.Run)

	fmt.Printf("PR #%d completed successfully\n", pr.Number)
}

func executePhase(name string, commands []string) {
	if len(commands) == 0 {
		return
	}
	fmt.Printf("Running %s commands...\n", name)
	if err := run.ExecuteCommands(commands, run.ShellExecutor); err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %v\n", name, err)
		os.Exit(1)
	}
}
