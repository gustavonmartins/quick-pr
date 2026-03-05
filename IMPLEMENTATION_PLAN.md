# Quick CI - Implementation Plan

## Project Status: Planning Complete, Ready to Implement

**Last Updated:** 2026-03-03
**Current Phase:** Planning
**Next Step:** Start with Step 1 (Config Parser)

---

## Project Overview

Quick CI is a lightweight CI/CD tool that:
- Monitors a Git repository for new pull requests
- Polls every 15 minutes for changes
- Downloads PRs and runs configured commands
- Stores results in JSON files (one per PR)
- Works with multiple Git providers (GitHub, GitLab, Bitbucket)

---

## Key Design Decisions

✅ **Multiple Git Providers:** Supported implicitly via URL detection
✅ **Authentication:** None required (public repos only for now)
✅ **PR Updates:** New commit to existing PR = treat as fresh, re-run tests
✅ **Results Storage:** One file per PR: `ci-results-pr-{number}.json`
✅ **Processing Mode:** Sequential (one PR at a time)
✅ **Failure Handling:** Fail fast (stop on first command failure)

---

## Configuration Format

**File:** `config.json`
```json
{
  "repository": "https://github.com/user/repo",
  "polling_interval_minutes": 15,
  "commands": ["go test ./...", "go build"],
  "results_directory": "./ci-results"
}
```

---

## Results Format

**File:** `ci-results/ci-results-pr-42.json`
```json
{
  "pr_number": 42,
  "pr_title": "Add new feature",
  "commit_sha": "abc123def456",
  "timestamp": "2026-03-03T10:15:00Z",
  "duration_seconds": 12.5,
  "status": "success",
  "commands": [
    {
      "command": "go test ./...",
      "exit_code": 0,
      "stdout": "PASS\nok ...",
      "stderr": "",
      "duration_seconds": 10.2
    },
    {
      "command": "go build",
      "exit_code": 0,
      "stdout": "",
      "stderr": "",
      "duration_seconds": 2.3
    }
  ]
}
```

---

## State Tracking Format

**File:** `.quick-ci-state.json`
```json
{
  "last_checked": "2026-03-03T10:00:00Z",
  "prs": {
    "42": "abc123def456",
    "43": "def456789abc"
  }
}
```

**Logic:**
- Key = PR number
- Value = last processed commit SHA
- If current SHA ≠ stored SHA → re-run (treat as fresh)
- If current SHA = stored SHA → skip (already processed)

---

## Implementation Steps

### ✅ Step 0: Project Setup
**Status:** COMPLETE
- [x] Create project directory
- [x] Initialize Go module
- [x] Create REQUIREMENTS.md
- [x] Create WORKFLOW.md
- [x] Create IMPLEMENTATION_PLAN.md

---

### ⏳ Step 1: Configuration File Parser
**Status:** NOT STARTED
**Dependencies:** None

**Implementation Details:**
- Define `Config` struct with fields:
  - `Repository` (string, required)
  - `PollingIntervalMinutes` (int, default: 15)
  - `Commands` ([]string, required)
  - `ResultsDirectory` (string, default: "./ci-results")
- Function: `LoadConfig(filepath string) (*Config, error)`
- Validate:
  - File exists and is valid JSON
  - Repository URL is not empty
  - At least one command specified
  - Polling interval > 0

**Test Coverage:**
- Valid config file loads successfully
- Missing required fields return error
- Invalid JSON returns error
- Default values applied correctly
- Invalid polling interval rejected

**Files to Create:**
- `config/config.go` - Config struct and LoadConfig function
- `config/config_test.go` - Tests

---

### ⏳ Step 2: Command Executor
**Status:** NOT STARTED
**Dependencies:** None (can be tested independently)

**Implementation Details:**
- Define `CommandResult` struct:
  - `Command` (string)
  - `ExitCode` (int)
  - `Stdout` (string)
  - `Stderr` (string)
  - `DurationSeconds` (float64)
- Function: `ExecuteCommands(commands []string, workDir string) ([]CommandResult, error)`
- Execute commands sequentially
- Stop on first non-zero exit code (fail fast)
- Capture all output
- Track duration per command

**Test Coverage:**
- Single command executes successfully
- Multiple commands execute in order
- Failed command stops execution (fail fast)
- Stdout/stderr captured correctly
- Duration tracked accurately
- Working directory respected

**Files to Create:**
- `executor/executor.go` - Command execution logic
- `executor/executor_test.go` - Tests

---

### ⏳ Step 3: Results Storage
**Status:** NOT STARTED
**Dependencies:** Step 2 (uses CommandResult)

**Implementation Details:**
- Define `TestResult` struct:
  - `PRNumber` (int)
  - `PRTitle` (string)
  - `CommitSHA` (string)
  - `Timestamp` (time.Time)
  - `DurationSeconds` (float64)
  - `Status` (string: "success" or "failure")
  - `Commands` ([]CommandResult)
- Function: `SaveResult(result *TestResult, directory string) error`
- Filename format: `ci-results-pr-{number}.json`
- Create results directory if doesn't exist
- Pretty-print JSON with indentation

**Test Coverage:**
- Result saved to correct filename
- JSON format is valid and readable
- Directory created if missing
- Existing file overwritten correctly
- Timestamp formatted correctly

**Files to Create:**
- `results/results.go` - Result struct and SaveResult function
- `results/results_test.go` - Tests

---

### ⏳ Step 4: Git Provider Detection
**Status:** NOT STARTED
**Dependencies:** None

**Implementation Details:**
- Function: `DetectProvider(repoURL string) (string, error)`
- Parse URL and detect provider:
  - `github.com` → "github"
  - `gitlab.com` → "gitlab"
  - `bitbucket.org` → "bitbucket"
- Return error for unsupported/invalid URLs
- Handle both HTTPS and SSH URLs

**Test Coverage:**
- GitHub URL detected correctly (HTTPS)
- GitLab URL detected correctly (HTTPS)
- Bitbucket URL detected correctly (HTTPS)
- Invalid URL returns error
- Unsupported provider returns error

**Files to Create:**
- `git/provider.go` - Provider detection
- `git/provider_test.go` - Tests

---

### ⏳ Step 5: Git Provider Client (GitHub First)
**Status:** NOT STARTED
**Dependencies:** Step 4 (provider detection)

**Implementation Details:**
- Define `PullRequest` struct:
  - `Number` (int)
  - `Title` (string)
  - `Branch` (string)
  - `CommitSHA` (string)
- Interface: `GitProvider`
  - `FetchPullRequests() ([]PullRequest, error)`
- Implement GitHub provider:
  - Use GitHub REST API v3
  - Endpoint: `GET /repos/{owner}/{repo}/pulls`
  - No authentication (public repos)
  - Parse response into PullRequest structs

**Test Coverage:**
- Can parse owner/repo from GitHub URL
- API request formed correctly
- Response parsed into PullRequest structs
- Empty PR list handled
- API errors handled gracefully

**Files to Create:**
- `git/client.go` - GitProvider interface
- `git/github.go` - GitHub implementation
- `git/github_test.go` - Tests

**Future Extensions:**
- `git/gitlab.go` - GitLab implementation
- `git/bitbucket.go` - Bitbucket implementation

---

### ⏳ Step 6: PR State Tracking
**Status:** NOT STARTED
**Dependencies:** Step 5 (uses PullRequest struct)

**Implementation Details:**
- Define `State` struct:
  - `LastChecked` (time.Time)
  - `PRs` (map[int]string) - PR number → commit SHA
- Function: `LoadState(filepath string) (*State, error)`
- Function: `SaveState(state *State, filepath string) error`
- Function: `ShouldProcess(pr PullRequest, state *State) bool`
  - Returns true if PR is new or commit SHA changed
  - Returns false if already processed with same SHA
- State file: `.quick-ci-state.json`

**Test Coverage:**
- New state file created with defaults
- Existing state loaded correctly
- State saved correctly
- New PR returns true for ShouldProcess
- Unchanged PR returns false for ShouldProcess
- Updated PR (new SHA) returns true for ShouldProcess

**Files to Create:**
- `state/state.go` - State management
- `state/state_test.go` - Tests

---

### ⏳ Step 7: PR Download/Checkout
**Status:** NOT STARTED
**Dependencies:** Step 5 (uses PullRequest struct)

**Implementation Details:**
- Function: `CloneAndCheckout(repoURL string, pr PullRequest) (workDir string, cleanup func(), error)`
- Clone repository to temp directory
- Fetch PR branch: `git fetch origin pull/{number}/head`
- Checkout specific commit: `git checkout {sha}`
- Return:
  - `workDir`: path to cloned repo
  - `cleanup`: function to remove temp directory
  - `error`: any error that occurred
- Cleanup function removes temp directory

**Test Coverage:**
- Repository cloned successfully
- PR branch fetched correctly
- Specific commit checked out
- Working directory path returned
- Cleanup removes temp directory
- Errors handled (invalid repo, network issues)

**Files to Create:**
- `git/clone.go` - Clone and checkout logic
- `git/clone_test.go` - Tests (may need mock/test repo)

---

### ⏳ Step 8: Polling Mechanism
**Status:** NOT STARTED
**Dependencies:** None (independent component)

**Implementation Details:**
- Function: `StartPolling(interval time.Duration, checkFunc func()) (stop func())`
- Run `checkFunc` on interval
- Return stop function for graceful shutdown
- Handle context cancellation
- Log each poll attempt

**Test Coverage:**
- Function called on interval
- Stop function cancels polling
- No goroutine leaks after stop
- Timing accuracy within reasonable bounds

**Files to Create:**
- `poller/poller.go` - Polling logic
- `poller/poller_test.go` - Tests

---

### ⏳ Step 9: Main Orchestration
**Status:** NOT STARTED
**Dependencies:** All previous steps

**Implementation Details:**
- Main function in `main.go`
- Load config
- Detect Git provider
- Create provider client
- Load state
- Start polling loop:
  - Fetch open PRs
  - For each PR (sequential):
    - Check if should process (new/updated)
    - Clone and checkout
    - Execute commands
    - Save results
    - Update state
    - Cleanup
- Handle signals (SIGINT, SIGTERM) for graceful shutdown
- CLI flags:
  - `--config` - config file path (default: `config.json`)
  - `--once` - run once and exit (no polling)

**Test Coverage:**
- Integration test: end-to-end with mock repo
- Config loading integrated correctly
- All components work together
- Graceful shutdown works
- State persists across runs

**Files to Create:**
- `main.go` - Main orchestration
- `main_test.go` - Integration tests

---

## Implementation Order Summary

1. ✅ **Step 0:** Project Setup (COMPLETE)
2. ⏳ **Step 1:** Config Parser → START HERE NEXT
3. ⏳ **Step 2:** Command Executor
4. ⏳ **Step 3:** Results Storage
5. ⏳ **Step 4:** Git Provider Detection
6. ⏳ **Step 5:** Git Provider Client (GitHub)
7. ⏳ **Step 6:** PR State Tracking
8. ⏳ **Step 7:** PR Download/Checkout
9. ⏳ **Step 8:** Polling Mechanism
10. ⏳ **Step 9:** Main Orchestration

---

## How to Resume

1. Read `WORKFLOW.md` to understand TDD process
2. Read this file to see current status
3. Find next step marked as "NOT STARTED"
4. Follow TDD workflow:
   - **Phase 0:** User requests feature → Claude suggests tests
   - **Phase 1:** Write tests (RED)
   - **Phase 2:** Implement until tests pass (GREEN)
   - **Phase 3:** Refactor if needed

**To start next session, say:**
"Let's implement Step 1: Configuration File Parser"

---

## Notes & Future Improvements

- **Step 5** currently only implements GitHub; GitLab and Bitbucket can be added later
- Authentication support can be added in future (tokens, SSH keys)
- Parallel PR processing can replace sequential in future
- Webhook support instead of polling (more efficient)
- Web UI for viewing results
- Notifications (email, Slack, etc.)
