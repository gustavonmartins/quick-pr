package download

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// TestFetchOpenPRsFromFeederRepo verifies that FetchPullRequests successfully fetches
// all open pull requests from a real GitHub repository (spacecowboy/Feeder).
func TestFetchOpenPRsFromFeederRepo(t *testing.T) {
	config, err := LoadConfig("../../testdata/feeder_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	prs, err := FetchPullRequests(config)
	if err != nil {
		t.Fatalf("Failed to fetch PRs: %v", err)
	}

	if len(prs) != 19 {
		t.Errorf("Expected 19 open PRs, got %d", len(prs))
	}
}

// TestFetchPR1030Details verifies that FetchPullRequests correctly fetches detailed
// information for a specific PR, including number, from/to branches, and commit count.
func TestFetchPR1030Details(t *testing.T) {
	config, err := LoadConfig("../../testdata/feeder_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expected, err := loadExpectedPR("../../testdata/pr_1030_expected.json")
	if err != nil {
		t.Fatalf("Failed to load expected output: %v", err)
	}

	prs, err := FetchPullRequests(config)
	if err != nil {
		t.Fatalf("Failed to fetch PRs: %v", err)
	}

	var pr1030 *PullRequest
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

// loadExpectedPR is a test helper that loads expected PR data from a JSON file.
func loadExpectedPR(filepath string) (*PullRequest, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var pr PullRequest
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

// TestSavePRsToOutputFolder verifies that SavePRsToFiles correctly saves all fetched
// pull requests as individual JSON files and that the saved data matches the original.
func TestSavePRsToOutputFolder(t *testing.T) {
	config, err := LoadConfig("../../testdata/feeder_config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	outputDir := "../../testdata/output"
	_ = os.RemoveAll(outputDir)

	prs, err := FetchPullRequests(config)
	if err != nil {
		t.Fatalf("Failed to fetch PRs: %v", err)
	}

	err = SavePRsToFiles(prs, outputDir)
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

		var savedPR PullRequest
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
