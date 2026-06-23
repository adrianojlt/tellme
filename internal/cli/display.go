package cli

import (
	"fmt"

	"tellme/internal/domain"
)

func PrintSuggestions(suggestions []domain.CommandSuggestion) {

	fmt.Println()

	for i, s := range suggestions {

		badges := ""

		if s.RiskLevel == "high" {
			badges += " [HIGH]"
		}

		if s.RequiresSudo {
			badges += " [sudo]"
		}

		fmt.Printf("%d. %s%s\n", i+1, s.Title, badges)
		fmt.Printf("   $ %s\n", s.Command)
		fmt.Printf("   %s\n", s.Description)
		fmt.Printf("   Risk: %s\n", s.RiskLevel)
		fmt.Println()
	}

	fmt.Println("0 - exit")
}

// PrintEntriesCompact prints suggestions in reverse order showing only the
// index and command - used for history/favorites/list views.
func PrintEntriesCompact(suggestions []domain.CommandSuggestion) {

	fmt.Println()

	for i := len(suggestions) - 1; i >= 0; i-- {
		fmt.Printf("%d. $ %s\n", i+1, suggestions[i].Command)
	}

	fmt.Println()
	fmt.Println("0 - exit")
}
