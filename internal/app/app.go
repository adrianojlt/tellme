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

// RunCommand executes command via the shell and records it in history. It does
// not interact with the user and never copies to the clipboard. Used by the
// favorites/list re-run flow when the user picked "Run" from the action menu.
func (a *App) RunCommand(command string) error {

	if err := a.runCommand(command); err != nil {
		return err
	}

	a.record("", "", command)

	return nil
}

// RunAndCopyCommand executes command via the shell, records it, and copies it
// to the clipboard. Used by the favorites/list flow for "Run and Copy".
func (a *App) RunAndCopyCommand(command string) error {

	if err := a.runCommand(command); err != nil {
		return err
	}

	if err := a.copyFn(command); err != nil {
		fmt.Printf("Clipboard error: %v\n", err)
	} else {
		fmt.Println("Copied to clipboard.")
	}

	a.record("", "", command)

	return nil
}

// CopyCommand copies command to the clipboard without executing it and records
// it in history. Used by the favorites/list flow for "Copy to clipboard".
func (a *App) CopyCommand(command string) error {

	if err := a.copyFn(command); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}

	fmt.Println("Copied to clipboard.")
	a.record("", "", command)

	return nil
}

func (a *App) runSelected(r io.Reader, query, title, command string) error {

	reader := bufio.NewReader(r)

	fmt.Print("Execute command? (y/n): ")

	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("reading input: %w", err)
	}

	if strings.TrimSpace(line) == "y" {

		if err := a.runCommand(command); err != nil {
			return err
		}

		a.record(query, title, command)
		return nil
	}

	if a.copyAfterSelect {

		if err := a.copyFn(command); err != nil {
			fmt.Printf("Clipboard error: %v\n", err)
			return nil
		}

		fmt.Println("Copied to clipboard.")
		a.record(query, title, command)
		return nil
	}

	fmt.Print("Copy to clipboard? (y/n): ")
	line, err = reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("reading input: %w", err)
	}

	if strings.TrimSpace(line) == "y" {
		if err := a.copyFn(command); err != nil {
			fmt.Printf("Clipboard error: %v\n", err)
			return nil
		}
		fmt.Println("Copied to clipboard.")
		a.record(query, title, command)
	}

	return nil
}

func (a *App) runCommand(command string) error {

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
