package domain

type CommandSuggestion struct {
	Title        string `json:"title"`
	Command      string `json:"command"`
	Description  string `json:"description"`
	RiskLevel    string `json:"risk_level"`
	RequiresSudo bool   `json:"requires_sudo"`
	Platform     string `json:"platform"`
}

type SuggestRequest struct {
	UserQuery  string
	MaxOptions int
	OSHint     string
	ShellHint  string
}
