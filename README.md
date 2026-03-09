# quick-ci

A lightweight CI tool that fetches GitHub PRs and runs commands against them.

## Usage

The tool operates in two phases:

### Phase 1: Download PRs (`-config`)

Fetches open PRs from a GitHub repository and generates JSON files with commands to execute.

```bash
go run main.go -config config.json -output ./output
```

| Flag | Description | Default |
|------|-------------|---------|
| `-config` | Path to config file | (required) |
| `-output` | Directory for PR JSON files | `./output` |

### Phase 2: Run Commands (`-run`)

Executes the commands from the generated PR JSON files.

```bash
go run main.go -run ./output
```

This runs four phases for each PR:
1. **setup** - One-time repository setup (skipped if workdir exists)
2. **per_pr** - Per-PR setup (checkout worktree)
3. **merge** - Merge strategy commands
4. **run** - Test/build commands

Results are written to `results.json` in the output directory.

## Configuration File

```json
{
  "repository": "https://github.com/owner/repo",
  "workdir": "./workdir",
  "merge_strategy": "merge",
  "setup": [
    "git clone --bare {repo} {workdir}/.git"
  ],
  "per_pr": [
    "git -C {workdir} worktree remove ./pr-{pr_number} || true",
    "git -C {workdir} fetch origin pull/{pr_number}/head",
    "git -C {workdir} worktree add ./pr-{pr_number} {sha}"
  ],
  "run": [
    "go test ./..."
  ],
  "polling_interval_minutes": 15
}
```

### Template Variables

| Variable | Description |
|----------|-------------|
| `{repo}` | GitHub repository URL |
| `{workdir}` | Working directory path |
| `{pr_number}` | Pull request number |
| `{sha}` | Commit SHA of PR head |

### Merge Strategies

| Strategy | Description |
|----------|-------------|
| `none` | No merge commands |
| `merge` | `git merge origin/{base}` |
| `rebase` | `git rebase origin/{base}` |
| `squash` | `git merge --squash origin/{base}` |

## Full Example

```bash
# Build the tool
go build -o quick-ci

# Download PRs
./quick-ci -config config.json -output ./prs

# Run tests on all PRs
./quick-ci -run ./prs
```
