package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"tellme/internal/domain"
)

func SelectSuggestion(suggestions []domain.CommandSuggestion) (*domain.CommandSuggestion, error) {
	return selectFrom(os.Stdin, suggestions)
}

func selectFrom(r io.Reader, suggestions []domain.CommandSuggestion) (*domain.CommandSuggestion, error) {

	reader := bufio.NewReader(r)

	for {
		fmt.Printf("Enter number (0-%d): ", len(suggestions))

		line, err := reader.ReadString('\n')

		if err == io.EOF {
			return nil, nil
		}

		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)

		n, parseErr := strconv.Atoi(line)

		if parseErr != nil || n < 0 || n > len(suggestions) {
			fmt.Printf("Invalid input. Enter a number between 0 and %d.\n", len(suggestions))
			continue
		}

		if n == 0 {
			return nil, nil
		}

		return &suggestions[n-1], nil
	}
}
