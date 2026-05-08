package cli

import (
	"fmt"

	"tellme/internal/domain"
)

func PrintSuggestions(suggestions []domain.CommandSuggestion) {
	fmt.Println()
	for i, s := range suggestions {
		fmt.Printf("%d. %s\n", i+1, s.Title)
		fmt.Printf("   $ %s\n", s.Command)
		fmt.Printf("   %s\n", s.Description)
		fmt.Printf("   Risk: %s\n", s.RiskLevel)
		fmt.Println()
	}
	fmt.Println("0 - exit")
}
