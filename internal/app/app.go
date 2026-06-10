package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"tellme/internal/domain"
	"tellme/internal/llm"
)

type App struct {
	provider        llm.Provider
	display         func([]domain.CommandSuggestion)
	selectFn        func([]domain.CommandSuggestion) (*domain.CommandSuggestion, error)
	copyFn          func(string) error
	maxOptions      int
	copyAfterSelect bool
	osHint          string
}

func New(
	p llm.Provider,
	display func([]domain.CommandSuggestion),
	selectFn func([]domain.CommandSuggestion) (*domain.CommandSuggestion, error),
	maxOptions int,
	copyFn func(string) error,
	copyAfterSelect bool,
	osHint string,
) *App {
	return &App{
		provider:        p,
		display:         display,
		selectFn:        selectFn,
		maxOptions:      maxOptions,
		copyFn:          copyFn,
		copyAfterSelect: copyAfterSelect,
		osHint:          osHint,
	}
}

func (a *App) Run(ctx context.Context, query string) error {
	suggestions, err := a.provider.SuggestCommands(ctx, domain.SuggestRequest{
		UserQuery:  query,
		MaxOptions: a.maxOptions,
		OSHint:     a.osHint,
	})
	if err != nil {
		return err
	}

	a.display(suggestions)

	selected, err := a.selectFn(suggestions)
	if err != nil {
		return err
	}

	if selected == nil {
		fmt.Println("Bye.")
		return nil
	}

	fmt.Printf("Selected: %s\n", selected.Command)
	a.handlePostSelect(selected.Command)

	return nil
}

func (a *App) handlePostSelect(command string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Execute command? (y/n): ")
	line, _ := reader.ReadString('\n')
	if strings.TrimSpace(line) == "y" {
		a.runCommand(command)
		return
	}

	if a.copyAfterSelect {
		if err := a.copyFn(command); err != nil {
			fmt.Printf("Clipboard error: %v\n", err)
		} else {
			fmt.Println("Copied to clipboard.")
		}
		return
	}

	fmt.Print("Copy to clipboard? (y/n): ")
	line, _ = reader.ReadString('\n')
	if strings.TrimSpace(line) == "y" {
		if err := a.copyFn(command); err != nil {
			fmt.Printf("Clipboard error: %v\n", err)
		} else {
			fmt.Println("Copied to clipboard.")
		}
	}
}

func (a *App) runCommand(command string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
