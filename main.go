package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/quick-ci/internal/common"
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

	var results []common.PRResult
	for _, file := range files {
		result := processPRFile(file)
		results = append(results, result)
	}

	// Write results to JSON file
	resultsFile := filepath.Join(dir, "results.json")
	writeResults(results, resultsFile)

	// Print summary
	printSummary(results, resultsFile)
}

func processPRFile(file string) common.PRResult {
	pr, err := run.LoadPRCommands(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", file, err)
		return common.PRResult{
			Success: false,
			Phases: []common.PhaseResult{{
				Name:    "load",
				Success: false,
				Commands: []common.CommandResult{{
					Command: "load " + file,
					Success: false,
					Output:  err.Error(),
				}},
			}},
		}
	}

	fmt.Printf("\n=== PR #%d: %s ===\n", pr.Number, pr.Title)

	result := common.PRResult{
		Number:  pr.Number,
		Title:   pr.Title,
		Success: true,
		Phases:  make([]common.PhaseResult, 0, 4),
	}

	phases := []struct {
		name     string
		commands []string
	}{
		{"setup", pr.Commands.Setup},
		{"per-pr", pr.Commands.PerPR},
		{"merge", pr.Commands.Merge},
		{"run", pr.Commands.Run},
	}

	for _, phase := range phases {
		phaseResult := run.ExecutePhase(phase.name, phase.commands)
		result.Phases = append(result.Phases, phaseResult)

		if !phaseResult.Success {
			result.Success = false
			fmt.Fprintf(os.Stderr, "PR #%d failed at %s phase\n", pr.Number, phase.name)
			break
		}
	}

	if result.Success {
		fmt.Printf("PR #%d completed successfully\n", pr.Number)
	}

	return result
}

func writeResults(results []common.PRResult, filename string) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling results: %v\n", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing results file: %v\n", err)
		return
	}
}

func printSummary(results []common.PRResult, resultsFile string) {
	var succeeded, failed int
	for _, r := range results {
		if r.Success {
			succeeded++
		} else {
			failed++
		}
	}

	fmt.Printf("\n========== SUMMARY ==========\n")
	fmt.Printf("Total: %d PRs\n", len(results))
	fmt.Printf("Succeeded: %d\n", succeeded)
	fmt.Printf("Failed: %d\n", failed)

	if failed > 0 {
		fmt.Printf("\nFailed PRs:\n")
		for _, r := range results {
			if !r.Success {
				failedPhase := findFailedPhase(r)
				fmt.Printf("  - PR #%d (%s): failed at %s\n", r.Number, r.Title, failedPhase)
			}
		}
	}

	fmt.Printf("\nResults written to: %s\n", resultsFile)

	if failed > 0 {
		os.Exit(1)
	}
}

func findFailedPhase(r common.PRResult) string {
	for _, phase := range r.Phases {
		if !phase.Success {
			return phase.Name
		}
	}
	return "unknown"
}
