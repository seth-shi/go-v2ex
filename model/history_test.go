package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestHistoryMissingFileAndAutomaticLoad(t *testing.T) {
	historyPath := filepath.Join(t.TempDir(), "history.json")
	store := NewHistoryStore(historyPath)
	if err := store.Load(); err != nil {
		t.Fatal(err)
	}
	if len(store.Entries()) != 0 || store.Contains(1) {
		t.Fatal("missing history must be empty")
	}
	if _, err := os.Stat(historyPath); !os.IsNotExist(err) {
		t.Fatal("loading missing history must not create a file")
	}
	if err := store.Record(HistoryEntry{TopicID: 1, Title: "first"}); err != nil {
		t.Fatal(err)
	}

	// A new store must retain existing history when its first operation is a
	// record, without relying on the UI to call Load first.
	reopened := NewHistoryStore(historyPath)
	before := time.Now().Unix()
	if err := reopened.Record(HistoryEntry{TopicID: 2, Title: "second"}); err != nil {
		t.Fatal(err)
	}
	entries := reopened.Entries()
	if len(entries) != 2 || entries[0].TopicID != 2 || entries[1].TopicID != 1 {
		t.Fatalf("existing history was not retained: %+v", entries)
	}
	if entries[0].VisitedAt < before || entries[0].VisitedAt > time.Now().Unix() {
		t.Fatal("record without visit time must receive the current time")
	}
	if !reopened.Contains(1) || reopened.Contains(3) {
		t.Fatal("Contains must reflect loaded entries")
	}
	entries[0].Title = "changed outside store"
	if reopened.Entries()[0].Title != "second" {
		t.Fatal("Entries exposed the store's backing slice")
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(historyPath)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Fatalf("history permissions = %o, want 600", info.Mode().Perm())
		}
	}
}

func TestHistoryRevisitMovesUpdatedEntryToFrontAndCapsSize(t *testing.T) {
	historyPath := filepath.Join(t.TempDir(), "history.json")
	entries := make([]HistoryEntry, maxHistoryEntries+5)
	for i := range entries {
		entries[i] = HistoryEntry{TopicID: int64(i + 1), Title: fmt.Sprint(i + 1), VisitedAt: 1}
	}
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(historyPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	store := NewHistoryStore(historyPath)
	if err := store.Record(HistoryEntry{TopicID: 3, Title: "updated", Node: "programmer", Author: "alice", URL: "https://www.v2ex.com/t/3", VisitedAt: 2}); err != nil {
		t.Fatal(err)
	}
	if err := store.Record(HistoryEntry{TopicID: 999, Title: "new", VisitedAt: 3}); err != nil {
		t.Fatal(err)
	}
	if err := store.Load(); err != nil {
		t.Fatal(err)
	}
	actual := store.Entries()
	if len(actual) != maxHistoryEntries || actual[0].TopicID != 999 || actual[1].TopicID != 3 || actual[1].Title != "updated" {
		t.Fatalf("unexpected recent history: length=%d, first=%+v", len(actual), actual[:2])
	}
	if actual[1].Node != "programmer" || actual[1].Author != "alice" || actual[1].URL != "https://www.v2ex.com/t/3" || actual[1].VisitedAt != 2 {
		t.Fatal("record metadata did not survive persistence")
	}
	if store.Contains(200) || store.Contains(205) {
		t.Fatal("oldest entries must be evicted")
	}
	seen := map[int64]bool{}
	for _, entry := range actual {
		if seen[entry.TopicID] {
			t.Fatalf("duplicate topic %d", entry.TopicID)
		}
		seen[entry.TopicID] = true
	}
}

func TestHistoryCorruptFileIsPreserved(t *testing.T) {
	for _, content := range []string{"{broken", "null", "{}", `[{"topic_id":0}]`, `[{"topic_id":1,"visited_at":-1}]`} {
		t.Run(content, func(t *testing.T) {
			historyPath := filepath.Join(t.TempDir(), "history.json")
			if err := os.WriteFile(historyPath, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			store := NewHistoryStore(historyPath)
			if err := store.Load(); err == nil {
				t.Fatal("corrupt file must return a load error")
			}
			if err := store.Record(HistoryEntry{TopicID: 2}); err == nil {
				t.Fatal("record must not overwrite corrupt history")
			}
			actual, err := os.ReadFile(historyPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(actual) != content {
				t.Fatal("corrupt history was modified")
			}
		})
	}
}

func TestHistoryFailedReloadPreventsOverwriteAndFailedSaveKeepsMemory(t *testing.T) {
	directory := t.TempDir()
	historyPath := filepath.Join(directory, "history.json")
	store := NewHistoryStore(historyPath)
	if err := store.Record(HistoryEntry{TopicID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(historyPath, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := store.Load(); err == nil {
		t.Fatal("reload must report corruption")
	}
	if err := store.Record(HistoryEntry{TopicID: 2}); err == nil {
		t.Fatal("a previously loaded store must not overwrite a failed reload")
	}
	if !store.Contains(1) || store.Contains(2) {
		t.Fatal("failed load or record must retain the previous in-memory view")
	}

	// Renaming a file onto a nonempty directory fails on supported platforms.
	// The temporary write must not become visible as committed history.
	if err := os.Remove(historyPath); err != nil {
		t.Fatal(err)
	}
	if err := store.Load(); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(historyPath, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(historyPath, "keep"), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := store.Record(HistoryEntry{TopicID: 3}); err == nil {
		t.Fatal("failed replacement must return an error")
	}
	if len(store.Entries()) != 0 {
		t.Fatal("failed save changed in-memory history")
	}
	files, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name() != "history.json" {
		t.Fatalf("temporary files were not cleaned up: %+v", files)
	}
}

func TestHistoryConcurrentVisitsAndReaders(t *testing.T) {
	historyPath := filepath.Join(t.TempDir(), "history.json")
	store := NewHistoryStore(historyPath)
	var group sync.WaitGroup
	for i := int64(1); i <= 24; i++ {
		group.Add(1)
		go func(id int64) {
			defer group.Done()
			if err := store.Record(HistoryEntry{TopicID: id}); err != nil {
				t.Errorf("record: %v", err)
			}
			store.Contains(id)
			store.Entries()
		}(i)
	}
	group.Wait()
	if err := store.Load(); err != nil {
		t.Fatal(err)
	}
	if len(store.Entries()) != 24 {
		t.Fatal("concurrent visits were lost")
	}
}

func TestHistoryRejectsInvalidVisitBeforeCreatingFile(t *testing.T) {
	historyPath := filepath.Join(t.TempDir(), "history.json")
	store := NewHistoryStore(historyPath)
	for _, entry := range []HistoryEntry{{}, {TopicID: -1}, {TopicID: 1, VisitedAt: -1}} {
		if err := store.Record(entry); err == nil {
			t.Fatal("invalid visit must fail")
		}
	}
	if _, err := os.Stat(historyPath); !os.IsNotExist(err) {
		t.Fatal("invalid visit must not create history")
	}
}
