package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	outputDir := flag.String("output", "./output", "directory to save PR JSON files")
	flag.Parse()

	config, err := LoadConfig(*configPath)
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
		err := SavePRWithCommands(pr, config, *outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving PR #%d: %v\n", pr.Number, err)
			os.Exit(1)
		}
		fmt.Printf("Saved pr-%d.json\n", pr.Number)
	}

	fmt.Printf("Done. Saved %d PR files to %s\n", len(prs), *outputDir)
}
