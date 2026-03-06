package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// httpClient is the HTTP client used for API requests.
// It can be replaced in tests to use a recording/caching transport.
var httpClient = http.DefaultClient

// PullRequest represents a pull request
type PullRequest struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	Head    Head   `json:"head"`
	Base    Base   `json:"base"`
	Commits int    `json:"commits"`
	From    string `json:"from"`
	To      string `json:"to"`
}

// Head represents the head branch information
type Head struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// Base represents the base branch information
type Base struct {
	Ref string `json:"ref"`
}

// FetchPullRequests fetches open pull requests from the repository
func FetchPullRequests(config *Config) ([]PullRequest, error) {
	owner, repo, err := parseGitHubURL(config.Repository)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=open", owner, repo)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PRs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Fetch full details for each PR to get commit count
	for i := range prs {
		fullPR, err := fetchSinglePR(owner, repo, prs[i].Number)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch PR #%d details: %w", prs[i].Number, err)
		}
		prs[i] = *fullPR
	}

	return prs, nil
}

// fetchSinglePR fetches detailed information for a single PR
func fetchSinglePR(owner, repo string, prNumber int) (*PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Populate From and To fields
	pr.From = pr.Head.Ref
	pr.To = pr.Base.Ref

	return &pr, nil
}

// parseGitHubURL extracts owner and repo from a GitHub URL
func parseGitHubURL(repoURL string) (owner, repo string, err error) {
	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Match github.com/owner/repo pattern
	re := regexp.MustCompile(`github\.com[:/]([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(repoURL)

	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid GitHub URL: %s", repoURL)
	}

	return matches[1], matches[2], nil
}

// SavePRsToFiles saves each PR as a separate JSON file in the output directory
func SavePRsToFiles(prs []PullRequest, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, pr := range prs {
		filename := filepath.Join(outputDir, fmt.Sprintf("pr-%d.json", pr.Number))
		data, err := json.MarshalIndent(pr, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal PR #%d: %w", pr.Number, err)
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("failed to write PR #%d: %w", pr.Number, err)
		}
	}

	return nil
}
