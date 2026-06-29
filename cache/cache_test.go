package cache_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/seraj/define/api"
	"github.com/seraj/define/cache"
)

func TestSetGet(t *testing.T) {
	dir := t.TempDir()

	s, err := cache.NewStore(filepath.Join(dir, "cache.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	entry := &cache.CacheEntry{
		Data: &api.Definition{
			Word:      "hello",
			Phonetics: []api.Phonetic{{Text: "/həˈloʊ/"}},
			Meanings: []api.Meaning{
				{
					PartOfSpeech: "interjection",
					Definitions: []api.Def{
						{Definition: "A greeting.", Example: "Hello, world!"},
					},
				},
			},
		},
		Preview:    "A greeting...",
		LookedUpAt: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
	}

	if err := s.Set("hello", entry); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, ok := s.Get("hello")
	if !ok {
		t.Fatal("Get returned false for existing key")
	}
	if got.Data.Word != "hello" {
		t.Errorf("word = %q, want %q", got.Data.Word, "hello")
	}
	if got.Preview != "A greeting..." {
		t.Errorf("preview = %q, want %q", got.Preview, "A greeting...")
	}
	if !got.LookedUpAt.Equal(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("LookedUpAt = %v", got.LookedUpAt)
	}

	// verify file was written
	raw, err := os.ReadFile(filepath.Join(dir, "cache.json"))
	if err != nil {
		t.Fatalf("cache file not written: %v", err)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, exists := data["hello"]; !exists {
		t.Error("key 'hello' not found in cache file")
	}
}

func TestGetMissing(t *testing.T) {
	dir := t.TempDir()
	s, err := cache.NewStore(filepath.Join(dir, "cache.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("Get should return false for missing key")
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	s, err := cache.NewStore(filepath.Join(dir, "cache.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	entry := &cache.CacheEntry{
		Data:       &api.Definition{Word: "test"},
		Preview:    "preview",
		LookedUpAt: time.Now(),
	}
	s.Set("test", entry)

	if err := s.Delete("test"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, ok := s.Get("test")
	if ok {
		t.Error("Get should return false after delete")
	}
}

func TestListSorted(t *testing.T) {
	dir := t.TempDir()
	s, err := cache.NewStore(filepath.Join(dir, "cache.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	t1 := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 30, 11, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 6, 30, 9, 0, 0, 0, time.UTC)

	s.Set("apple", &cache.CacheEntry{Data: &api.Definition{Word: "apple"}, Preview: "a fruit", LookedUpAt: t1})
	s.Set("banana", &cache.CacheEntry{Data: &api.Definition{Word: "banana"}, Preview: "yellow fruit", LookedUpAt: t2})
	s.Set("cherry", &cache.CacheEntry{Data: &api.Definition{Word: "cherry"}, Preview: "red fruit", LookedUpAt: t3})

	items, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("List count = %d, want 3", len(items))
	}
	if items[0].Word != "banana" {
		t.Errorf("most recent = %q, want banana", items[0].Word)
	}
	if items[1].Word != "apple" {
		t.Errorf("second = %q, want apple", items[1].Word)
	}
	if items[2].Word != "cherry" {
		t.Errorf("oldest = %q, want cherry", items[2].Word)
	}
}

func TestLowercaseKeys(t *testing.T) {
	dir := t.TempDir()
	s, err := cache.NewStore(filepath.Join(dir, "cache.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	entry := &cache.CacheEntry{
		Data:       &api.Definition{Word: "Hello"},
		Preview:    "greeting",
		LookedUpAt: time.Now(),
	}
	s.Set("Hello", entry)

	_, ok := s.Get("hello")
	if !ok {
		t.Error("Get lowercase should find uppercase entry")
	}

	_, ok = s.Get("HELLO")
	if !ok {
		t.Error("Get uppercase should find uppercase entry")
	}
}

func TestNewStoreCreatesDir(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "newsub", "cache.json")

	s, err := cache.NewStore(subDir)
	if err != nil {
		t.Fatalf("NewStore should create dirs: %v", err)
	}

	entry := &cache.CacheEntry{
		Data:       &api.Definition{Word: "test"},
		Preview:    "preview",
		LookedUpAt: time.Now(),
	}
	if err := s.Set("test", entry); err != nil {
		t.Fatalf("Set should work after auto-creating dir: %v", err)
	}
}
