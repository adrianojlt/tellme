package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tellme/internal/app"
	"tellme/internal/clipboard"
	"tellme/internal/config"
	"tellme/internal/domain"
	"tellme/internal/providers"
	"tellme/internal/store"
)

// Subcommand keywords that dispatch to history/favorites listing.
const (
	subHistory   = "history"
	subFavorites = "favorites"
)

// IsSubcommand reports whether args is a positional subcommand invocation,
// i.e. exactly one argument that exactly equals a known keyword. A query like
// "find big files" is not a subcommand because its first token is not an exact
// keyword match.
func IsSubcommand(args []string) bool {
	return len(args) == 1 && (args[0] == subHistory || args[0] == subFavorites)
}

// RunHistory lists the newest MaxHistory history entries and lets the user
// re-run or copy one without calling the LLM.
func RunHistory(cfgPath string) error {

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	storePath := storePathFor(cfgPath)
	s, err := store.Load(storePath)
	if err != nil {
		return err
	}

	entries := s.RecentHistory(cfg.Behavior.MaxHistory)
	return runHistoryListing(os.Stdin, s, storePath, entries, reRunner(cfg, storePath))
}

// runHistoryListing renders history entries, lets the user select one, then
// presents an action choice: re-run/copy (existing behavior), add to Favorites,
// or add to a List. It keeps the underlying []store.Entry so the chosen index
// maps back to the full Entry needed by store.AddToList. The reader and store
// are injected so the action flow is unit-testable. An empty list prints a
// friendly message and returns nil.
func runHistoryListing(r io.Reader, s *store.Store, storePath string, entries []store.Entry, reRun func(command string) error) error {

	if len(entries) == 0 {
		fmt.Println("No history yet.")
		return nil
	}

	suggestions := make([]domain.CommandSuggestion, len(entries))
	for i, e := range entries {
		suggestions[i] = e.ToSuggestion()
	}

	PrintEntriesCompact(suggestions)

	reader := bufio.NewReader(r)

	idx, err := selectIndex(reader, len(suggestions))
	if err != nil {
		return err
	}
	if idx < 0 {
		fmt.Println("Bye.")
		return nil
	}

	entry := entries[idx]
	fmt.Printf("Selected: %s\n", entry.Command)

	action, err := selectAction(reader)
	if err != nil {
		return err
	}

	switch action {
	case actionAddFavorites:
		return addToListAndSave(s, storePath, "favorites", entry)
	case actionAddList:
		return addToChosenList(reader, s, storePath, entry)
	default:
		return reRun(entry.Command)
	}
}

// RunFavorites lists the "favorites" list and lets the user re-run or copy one
// without calling the LLM.
func RunFavorites(cfgPath string) error {

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	storePath := storePathFor(cfgPath)
	s, err := store.Load(storePath)
	if err != nil {
		return err
	}

	actions := listRunnerActions(cfg, storePath)
	actions.Remove = func(command string) error {
		s.RemoveFromList("favorites", command)
		return store.Save(storePath, s)
	}
	return runListing(os.Stdin, s.Lists["favorites"], "favorites", actions)
}

// storePathFor derives the store path from the config path, mirroring main.go
// (~/.config/tellme/store.json).
func storePathFor(cfgPath string) string {
	return filepath.Join(filepath.Dir(cfgPath), "store.json")
}

// ListActions holds the four callbacks the favorites/list action menu can
// invoke. Run, RunAndCopy, and Copy are the three command-handling actions;
// Remove is the list-management action that mutates the store.
type ListActions struct {
	Run        func(command string) error
	RunAndCopy func(command string) error
	Copy       func(command string) error
	Remove     func(command string) error
}

// runListing renders entries, lets the user select one, then shows the
// favorites/list action menu (Run, Run and Copy, Copy, Remove). An empty list
// prints a friendly message and returns nil. All four callbacks are injected
// so the dispatch logic is unit-testable.
func runListing(r io.Reader, entries []store.Entry, label string, actions ListActions) error {

	if len(entries) == 0 {
		fmt.Printf("No %s yet.\n", label)
		return nil
	}

	suggestions := make([]domain.CommandSuggestion, len(entries))
	for i, e := range entries {
		suggestions[i] = e.ToSuggestion()
	}

	PrintEntriesCompact(suggestions)

	reader := bufio.NewReader(r)

	idx, err := selectIndex(reader, len(suggestions))
	if err != nil {
		return err
	}
	if idx < 0 {
		fmt.Println("Bye.")
		return nil
	}

	selected := suggestions[idx]
	fmt.Printf("Selected: %s\n", selected.Command)

	action, err := selectListAction(reader, label)
	if err != nil {
		return err
	}

	switch action {
	case actionListRemove:
		if err := actions.Remove(selected.Command); err != nil {
			return err
		}
		fmt.Printf("Removed from %s.\n", label)
		return nil
	case actionListRunAndCopy:
		return actions.RunAndCopy(selected.Command)
	case actionListCopy:
		return actions.Copy(selected.Command)
	default:
		return actions.Run(selected.Command)
	}
}

func selectListAction(reader *bufio.Reader, label string) (int, error) {
	fmt.Println("Action:")
	fmt.Println("  1) Run")
	fmt.Println("  2) Run and Copy to clipboard")
	fmt.Println("  3) Copy to clipboard")
	fmt.Printf("  4) Remove from %s\n", label)

	idx, err := selectIndex(reader, 4)
	if err != nil {
		return actionListRun, err
	}

	switch idx {
	case 1:
		return actionListRunAndCopy, nil
	case 2:
		return actionListCopy, nil
	case 3:
		return actionListRemove, nil
	default:
		return actionListRun, nil
	}
}

// Action choices offered after selecting a list/favorites item.
const (
	actionListRun = iota
	actionListRunAndCopy
	actionListCopy
	actionListRemove
)

// Action choices offered after selecting a history item.
const (
	actionRunCopy = iota
	actionAddFavorites
	actionAddList
)

// selectAction prompts the user to choose what to do with the selected history
// item. It returns one of the action* constants. An invalid or empty/EOF input
// defaults to actionRunCopy (the existing run/copy behavior).
func selectAction(reader *bufio.Reader) (int, error) {

	fmt.Println("Action:")
	fmt.Println("  1) Run / copy")
	fmt.Println("  2) Add to Favorites")
	fmt.Println("  3) Add to a List")

	idx, err := selectIndex(reader, 3)
	if err != nil {
		return actionRunCopy, err
	}

	switch idx {
	case 1:
		return actionAddFavorites, nil
	case 2:
		return actionAddList, nil
	default:
		return actionRunCopy, nil
	}
}

// addToChosenList shows existing list names numbered plus a "create new"
// option. Selecting an existing list adds the entry to it. "create new" prompts
// for a name (trimmed; empty/whitespace is rejected with a message and no list
// is created); if the typed name already exists it is reused. On a valid choice
// the entry is added and the store saved.
func addToChosenList(reader *bufio.Reader, s *store.Store, storePath string, entry store.Entry) error {

	names := make([]string, 0, len(s.Lists))
	for name := range s.Lists {
		names = append(names, name)
	}

	sort.Strings(names)

	fmt.Println("Add to list:")
	for i, name := range names {
		fmt.Printf("  %d) %s (%d)\n", i+1, name, len(s.Lists[name]))
	}

	fmt.Printf("  %d) Create new list\n", len(names)+1)

	idx, err := selectIndex(reader, len(names)+1)
	if err != nil {
		return err
	}

	if idx < 0 {
		fmt.Println("Bye.")
		return nil
	}

	if idx < len(names) {
		return addToListAndSave(s, storePath, names[idx], entry)
	}

	name, err := promptNewListName(reader)
	if err != nil {
		return err
	}

	if name == "" {
		fmt.Println("List name cannot be empty.")
		return nil
	}

	return addToListAndSave(s, storePath, name, entry)
}

// promptNewListName reads and trims a new list name. An empty/whitespace name
// (or EOF) returns "" so the caller can reject it.
func promptNewListName(reader *bufio.Reader) (string, error) {

	fmt.Print("New list name: ")

	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

// addToListAndSave adds entry to the named list and saves the store. A duplicate
// command (AddToList error) is reported with a message and leaves the list
// unchanged; the store is not saved in that case.
func addToListAndSave(s *store.Store, storePath, name string, entry store.Entry) error {

	if err := s.AddToList(name, entry); err != nil {
		fmt.Println(err)
		return nil
	}

	if err := store.Save(storePath, s); err != nil {
		return err
	}

	fmt.Printf("Added to %s.\n", name)
	return nil
}

// reRunner returns a function that builds a minimal App and runs the chosen
// command through the execute/copy flow. App construction mirrors main.go (real
// provider built from cfg) so re-run uses the same copyFn and behavior
// settings. The provider is never called during re-run. It is built lazily, so
// listing an empty or exit-selected view never requires a configured instance.
func reRunner(cfg *config.Config, storePath string) func(string) error {

	return func(command string) error {

		p, err := providers.New(cfg)

		if err != nil {
			return err
		}

		a := app.New(
			p,
			PrintSuggestions,
			SelectSuggestion,
			cfg.Behavior.MaxOptions,
			clipboard.Copy,
			cfg.Behavior.CopyAfterSelect,
			cfg.OS,
			storePath,
		)

		return a.RunSelected(command)
	}
}

// listRunnerActions builds the four-callback ListActions used by the
// favorites/list action menu. Each callback builds a fresh App (mirroring
// reRunner) and dispatches to the matching non-interactive App method. The
// Remove callback is left nil; callers must wire it to the list-specific store
// mutation before passing the struct to runListing.
func listRunnerActions(cfg *config.Config, storePath string) ListActions {

	buildApp := func() (*app.App, error) {

		p, err := providers.New(cfg)
		if err != nil {
			return nil, err
		}

		return app.New(
			p,
			PrintSuggestions,
			SelectSuggestion,
			cfg.Behavior.MaxOptions,
			clipboard.Copy,
			cfg.Behavior.CopyAfterSelect,
			cfg.OS,
			storePath,
		), nil
	}

	return ListActions{
		Run: func(command string) error {
			a, err := buildApp()
			if err != nil {
				return err
			}
			return a.RunCommand(command)
		},
		RunAndCopy: func(command string) error {
			a, err := buildApp()
			if err != nil {
				return err
			}
			return a.RunAndCopyCommand(command)
		},
		Copy: func(command string) error {
			a, err := buildApp()
			if err != nil {
				return err
			}
			return a.CopyCommand(command)
		},
	}
}
