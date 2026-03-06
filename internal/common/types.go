package common

// Head represents the head (source) branch of a PR
type Head struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// Base represents the base (target) branch of a PR
type Base struct {
	Ref string `json:"ref"`
}

// ParsedCommands holds the filled-in command templates
type ParsedCommands struct {
	Setup []string `json:"setup"`
	PerPR []string `json:"per_pr"`
	Merge []string `json:"merge"`
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

// CommandResult holds the result of a single command execution
type CommandResult struct {
	Command string `json:"command"`
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
}

// PhaseResult holds the result of running a phase (setup, per-PR, merge, run)
type PhaseResult struct {
	Name     string          `json:"name"`
	Success  bool            `json:"success"`
	Commands []CommandResult `json:"commands"`
}

// PRResult holds the result of running commands for a PR
type PRResult struct {
	Number  int           `json:"number"`
	Title   string        `json:"title"`
	Success bool          `json:"success"`
	Phases  []PhaseResult `json:"phases"`
}
