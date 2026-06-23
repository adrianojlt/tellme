package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"tellme/internal/domain"
	"tellme/internal/llm"
	"tellme/internal/store"
)

type App struct {
	provider        llm.Provider
	display         func([]domain.CommandSuggestion)
	selectFn        func([]domain.CommandSuggestion) (*domain.CommandSuggestion, error)
	copyFn          func(string) error
	maxOptions      int
	copyAfterSelect bool
	osHint          string
	storePath       string
}

func New(
	p llm.Provider,
	display func([]domain.CommandSuggestion),
	selectFn func([]domain.CommandSuggestion) (*domain.CommandSuggestion, error),
	maxOptions int,
	copyFn func(string) error,
	copyAfterSelect bool,
	osHint string,
	storePath string,
) *App {
	return &App{
		provider:        p,
		display:         display,
		selectFn:        selectFn,
		maxOptions:      maxOptions,
		copyFn:          copyFn,
		copyAfterSelect: copyAfterSelect,
		osHint:          osHint,
		storePath:       storePath,
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
	return a.runSelected(os.Stdin, query, selected.Title, selected.Command)
}

func (a *App) RunSelected(command string) error {
	return a.runSelected(os.Stdin, "", "", command)
}

func (a *App) runSelected(r io.Reader, query, title, command string) error {
	reader := bufio.NewReader(r)

	fmt.Print("Execute command? (y/n): ")
	line, _ := reader.ReadString('\n')
	if strings.TrimSpace(line) == "y" {
		a.runCommand(command)
		a.record(query, title, command)
		return nil
	}

	if a.copyAfterSelect {
		if err := a.copyFn(command); err != nil {
			fmt.Printf("Clipboard error: %v\n", err)
		} else {
			fmt.Println("Copied to clipboard.")
		}
		return nil
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

	return nil
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

// record appends the executed command to the history store. A store load/save
// error is reported to stderr but is non-fatal: it never changes the executed
// command's outcome. An empty storePath disables recording.
func (a *App) record(query, title, command string) {

	if a.storePath == "" {
		return
	}

	s, err := store.Load(a.storePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: %v\n", err)
		return
	}

	s.AddHistory(store.Entry{
		Query:     query,
		Command:   command,
		Title:     title,
		Timestamp: time.Now().UTC(),
	})

	if err := store.Save(a.storePath, s); err != nil {
		fmt.Fprintf(os.Stderr, "history: %v\n", err)
	}
}
