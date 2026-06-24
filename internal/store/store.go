package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tellme/internal/domain"
)

const maxHistory = 500

type Entry struct {
	Query     string    `json:"query"`
	Command   string    `json:"command"`
	Title     string    `json:"title"`
	Timestamp time.Time `json:"timestamp"`
}

type Store struct {
	History []Entry            `json:"history"`
	Lists   map[string][]Entry `json:"lists"`
}

func empty() *Store {
	return &Store{
		History: nil,
		Lists:   map[string][]Entry{},
	}
}

func Load(path string) (*Store, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return empty(), nil
		}
		return nil, err
	}

	s := empty()
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}

	if s.Lists == nil {
		s.Lists = map[string][]Entry{}
	}

	return s, nil
}

// AddHistory inserts e at the front of the history (newest-first). If an entry
// with the same Command already exists, it is removed and re-inserted at the
// front (bumped to top). History is capped at maxHistory, dropping oldest.
func (s *Store) AddHistory(e Entry) {

	// [:0:0] creates a nil-capacity slice sharing no backing array, so appends
	// below never mutate the original History slice.
	filtered := s.History[:0:0]
	for _, h := range s.History {
		if h.Command != e.Command {
			filtered = append(filtered, h)
		}
	}

	s.History = append([]Entry{e}, filtered...)

	if len(s.History) > maxHistory {
		s.History = s.History[:maxHistory]
	}
}

// AddToList appends e to the named list, creating it if absent. It returns a
// non-nil error if an entry with the same Command already exists in that list.
func (s *Store) AddToList(name string, e Entry) error {

	if s.Lists == nil {
		s.Lists = map[string][]Entry{}
	}

	for _, existing := range s.Lists[name] {
		if existing.Command == e.Command {
			return fmt.Errorf("command %q already in list %q", e.Command, name)
		}
	}

	s.Lists[name] = append(s.Lists[name], e)
	return nil
}

// RemoveFromList removes the entry with the given command from the named list.
// If the list becomes empty it is deleted. No-op if the command is not found.
func (s *Store) RemoveFromList(name, command string) {

	entries := s.Lists[name]
	// [:0:0] - see AddHistory for rationale.
	filtered := entries[:0:0]

	for _, e := range entries {
		if e.Command != command {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == 0 {
		delete(s.Lists, name)
	} else {
		s.Lists[name] = filtered
	}
}

// RecentHistory returns the newest n entries, or all of them if fewer than n
// exist. A non-positive n returns an empty slice.
func (s *Store) RecentHistory(n int) []Entry {

	if n <= 0 {
		return []Entry{}
	}

	if n > len(s.History) {
		n = len(s.History)
	}

	out := make([]Entry, n)
	copy(out, s.History[:n])

	return out
}

// ToSuggestion maps an Entry to a domain.CommandSuggestion, populating only
// Title and Command. Description, RiskLevel, RequiresSudo, and Platform are
// left empty.
func (e Entry) ToSuggestion() domain.CommandSuggestion {
	return domain.CommandSuggestion{
		Title:   e.Title,
		Command: e.Command,
	}
}

// FromSuggestion maps a domain.CommandSuggestion to an Entry, populating only
// Title and Command. The caller sets Query and Timestamp.
func FromSuggestion(s domain.CommandSuggestion) Entry {
	return Entry{
		Title:   s.Title,
		Command: s.Command,
	}
}

func Save(path string, s *Store) error {

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	f, err := os.CreateTemp(dir, ".store-*.tmp")
	if err != nil {
		return err
	}

	tmp := f.Name()

	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}

	return nil
}
