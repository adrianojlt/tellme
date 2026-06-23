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

	idx, err := selectIndex(bufio.NewReader(r), len(suggestions))

	if err != nil {
		return nil, err
	}

	if idx < 0 {
		return nil, nil
	}

	return &suggestions[idx], nil
}

// selectIndex prompts for a number in [0, count] and returns the zero-based
// index of the chosen item. A 0 selection or EOF returns -1 (exit/none). It
// takes a *bufio.Reader so callers that issue several prompts in sequence share
// one buffered reader (bufio reads ahead, so a fresh reader per prompt would
// drop already-buffered input).
func selectIndex(reader *bufio.Reader, count int) (int, error) {
	for {
		fmt.Printf("Enter number (0-%d): ", count)

		line, err := reader.ReadString('\n')

		if err == io.EOF {
			return -1, nil
		}

		if err != nil {
			return -1, err
		}

		line = strings.TrimSpace(line)

		n, parseErr := strconv.Atoi(line)

		if parseErr != nil || n < 0 || n > count {
			fmt.Printf("Invalid input. Enter a number between 0 and %d.\n", count)
			continue
		}

		if n == 0 {
			return -1, nil
		}

		return n - 1, nil
	}
}
