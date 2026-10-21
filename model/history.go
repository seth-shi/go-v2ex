package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const maxHistoryEntries = 200

// HistoryEntry describes a successfully opened topic, without account data.
type HistoryEntry struct {
	TopicID   int64  `json:"topic_id"`
	Title     string `json:"title"`
	Node      string `json:"node"`
	Author    string `json:"author"`
	URL       string `json:"url"`
	VisitedAt int64  `json:"visited_at"`
}

// HistoryStore keeps the most recently visited topics first. A store is safe
// for concurrent use; callers must load it before reading existing history.
type HistoryStore struct {
	mu      sync.RWMutex
	path    string
	loaded  bool
	entries []HistoryEntry
}

func NewHistoryStore(path string) *HistoryStore {
	return &HistoryStore{path: path}
}

func HistoryPath() string {
	return filepath.Join(filepath.Dir(ConfigPath()), ".go-v2ex-history.json")
}

// Load reads history without rewriting it. An unreadable or invalid file is
// retained, and must be successfully loaded before a subsequent Record.
func (s *HistoryStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *HistoryStore) loadLocked() error {
	s.loaded = false
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.entries = nil
		s.loaded = true
		return nil
	}
	if err != nil {
		return fmt.Errorf("read browsing history: %w", err)
	}

	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("decode browsing history: %w", err)
	}
	if entries == nil {
		return errors.New("decode browsing history: expected an array")
	}
	for _, entry := range entries {
		if entry.TopicID <= 0 || entry.VisitedAt < 0 {
			return errors.New("decode browsing history: invalid topic or visit time")
		}
	}

	seen := make(map[int64]bool, min(len(entries), maxHistoryEntries))
	retained := make([]HistoryEntry, 0, min(len(entries), maxHistoryEntries))
	for _, entry := range entries {
		if seen[entry.TopicID] {
			continue
		}
		seen[entry.TopicID] = true
		retained = append(retained, entry)
		if len(retained) == maxHistoryEntries {
			break
		}
	}
	s.entries = retained
	s.loaded = true
	return nil
}

// Entries returns a copy so views cannot mutate the store's state.
func (s *HistoryStore) Entries() []HistoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]HistoryEntry(nil), s.entries...)
}

func (s *HistoryStore) Contains(topicID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, entry := range s.entries {
		if entry.TopicID == topicID {
			return true
		}
	}
	return false
}

// Record should be called only after the topic was successfully fetched. It
// persists the new order before publishing it to readers in memory.
func (s *HistoryStore) Record(entry HistoryEntry) error {
	if entry.TopicID <= 0 {
		return errors.New("record browsing history: topic ID must be positive")
	}
	if entry.VisitedAt < 0 {
		return errors.New("record browsing history: visit time must not be negative")
	}
	if entry.VisitedAt == 0 {
		entry.VisitedAt = time.Now().Unix()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.loaded {
		if err := s.loadLocked(); err != nil {
			return err
		}
	}

	entries := make([]HistoryEntry, 1, min(len(s.entries)+1, maxHistoryEntries))
	entries[0] = entry
	for _, previous := range s.entries {
		if previous.TopicID == entry.TopicID {
			continue
		}
		if len(entries) == maxHistoryEntries {
			break
		}
		entries = append(entries, previous)
	}
	if err := s.writeLocked(entries); err != nil {
		return err
	}
	s.entries = entries
	return nil
}

func (s *HistoryStore) writeLocked(entries []HistoryEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encode browsing history: %w", err)
	}
	data = append(data, '\n')

	// A temporary file in the destination directory avoids exposing partial
	// JSON and keeps replacement on the same filesystem. Never delete the old
	// history first: a failed replacement must leave it available.
	file, err := os.CreateTemp(filepath.Dir(s.path), ".go-v2ex-history-*")
	if err != nil {
		return fmt.Errorf("create browsing history: %w", err)
	}
	temporaryPath := file.Name()
	defer os.Remove(temporaryPath)
	defer file.Close()
	if err := file.Chmod(0600); err != nil {
		return fmt.Errorf("protect browsing history: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write browsing history: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync browsing history: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close browsing history: %w", err)
	}
	if err := os.Rename(temporaryPath, s.path); err != nil {
		return fmt.Errorf("replace browsing history: %w", err)
	}
	return nil
}
