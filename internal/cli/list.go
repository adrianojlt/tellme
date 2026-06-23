package cli

import (
	"fmt"
	"os"
	"sort"

	"tellme/internal/config"
	"tellme/internal/store"
)

// subList is the keyword that dispatches to the list subcommands.
const subList = "list"

// IsListNames reports whether args is the bare "list" subcommand (no name),
// i.e. exactly one argument equal to the "list" keyword.
func IsListNames(args []string) bool {
	return len(args) == 1 && args[0] == subList
}

// IsListNamed reports whether args is the "list <name>" subcommand, i.e. exactly
// two arguments whose first token is the "list" keyword. A 3+ token "list ..."
// invocation is not a subcommand and falls through to query handling.
func IsListNamed(args []string) bool {
	return len(args) == 2 && args[0] == subList
}

// RunListNames prints every list name with its item count. The "favorites" list
// is included if present. When no lists exist, it prints a friendly message.
func RunListNames(cfgPath string) error {

	s, err := store.Load(storePathFor(cfgPath))

	if err != nil {
		return err
	}

	return printListNames(s.Lists)
}

// printListNames is the testable core of RunListNames: it renders names with
// counts (sorted for stable output) or a friendly message when empty.
func printListNames(lists map[string][]store.Entry) error {

	if len(lists) == 0 {
		fmt.Println("No lists yet.")
		return nil
	}

	names := make([]string, 0, len(lists))
	for name := range lists {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		fmt.Printf("%s (%d)\n", name, len(lists[name]))
	}

	return nil
}

// RunList displays the named list's items and lets the user re-run or copy one
// without calling the LLM. If the list is absent, it prints a "No such list"
// message and returns nil (exit 0).
func RunList(cfgPath, name string) error {

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	storePath := storePathFor(cfgPath)
	s, err := store.Load(storePath)
	if err != nil {
		return err
	}

	removeFn := func(command string) error {
		s.RemoveFromList(name, command)
		return store.Save(storePath, s)
	}

	return lookupList(s.Lists, name, func() error {
		return runListing(os.Stdin, s.Lists[name], name, reRunner(cfg, storePath), removeFn)
	})
}

// lookupList runs show when the named list exists; otherwise it prints a
// "No such list" message and returns nil so the process exits 0.
func lookupList(lists map[string][]store.Entry, name string, show func() error) error {

	if _, ok := lists[name]; !ok {
		fmt.Printf("No such list: %s\n", name)
		return nil
	}

	return show()
}
