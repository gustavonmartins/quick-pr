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
