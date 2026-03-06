package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetching PRs from %s...\n", config.Repository)
	prs, err := FetchPullRequests(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching PRs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d open PRs\n", len(prs))

	for _, pr := range prs {
		err := SavePRWithCommands(pr, config, outputDir)
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
		pr, err := LoadPRCommands(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", file, err)
			os.Exit(1)
		}

		fmt.Printf("\n=== PR #%d: %s ===\n", pr.Number, pr.Title)

		// Run setup commands
		if len(pr.Commands.Setup) > 0 {
			fmt.Println("Running setup commands...")
			if err := ExecuteCommands(pr.Commands.Setup, ShellExecutor); err != nil {
				fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
				os.Exit(1)
			}
		}

		// Run per-PR commands
		if len(pr.Commands.PerPR) > 0 {
			fmt.Println("Running per-PR commands...")
			if err := ExecuteCommands(pr.Commands.PerPR, ShellExecutor); err != nil {
				fmt.Fprintf(os.Stderr, "Per-PR commands failed: %v\n", err)
				os.Exit(1)
			}
		}

		// Run merge commands
		if len(pr.Commands.Merge) > 0 {
			fmt.Println("Running merge commands...")
			if err := ExecuteCommands(pr.Commands.Merge, ShellExecutor); err != nil {
				fmt.Fprintf(os.Stderr, "Merge commands failed: %v\n", err)
				os.Exit(1)
			}
		}

		// Run CI commands
		if len(pr.Commands.Run) > 0 {
			fmt.Println("Running CI commands...")
			if err := ExecuteCommands(pr.Commands.Run, ShellExecutor); err != nil {
				fmt.Fprintf(os.Stderr, "CI commands failed: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("PR #%d completed successfully\n", pr.Number)
	}

	fmt.Printf("\nDone. Processed %d PRs\n", len(files))
}
