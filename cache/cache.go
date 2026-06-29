package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/seraj/define/api"
)

type CacheEntry struct {
	Data       *api.Definition `json:"data"`
	Preview    string          `json:"preview"`
	LookedUpAt time.Time       `json:"looked_up_at"`
}

type HistoryItem struct {
	Word       string
	Preview    string
	LookedUpAt time.Time
}

type Store struct {
	path string
}

func NewStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}
	return &Store{path: path}, nil
}

func (s *Store) Get(word string) (*CacheEntry, bool) {
	all, err := s.readAll()
	if err != nil {
		return nil, false
	}
	key := strings.ToLower(word)
	entry, ok := all[key]
	return entry, ok
}

func (s *Store) Set(word string, entry *CacheEntry) error {
	all, _ := s.readAll()
	if all == nil {
		all = make(map[string]*CacheEntry)
	}
	key := strings.ToLower(word)
	all[key] = entry
	return s.writeAll(all)
}

func (s *Store) Delete(word string) error {
	all, _ := s.readAll()
	if all == nil {
		return nil
	}
	key := strings.ToLower(word)
	delete(all, key)
	return s.writeAll(all)
}

func (s *Store) List() ([]HistoryItem, error) {
	all, err := s.readAll()
	if err != nil {
		return nil, err
	}
	var items []HistoryItem
	for word, entry := range all {
		items = append(items, HistoryItem{
			Word:       word,
			Preview:    entry.Preview,
			LookedUpAt: entry.LookedUpAt,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].LookedUpAt.After(items[j].LookedUpAt)
	})
	return items, nil
}

func (s *Store) readAll() (map[string]*CacheEntry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var all map[string]*CacheEntry
	if err := json.Unmarshal(data, &all); err != nil {
		return nil, fmt.Errorf("decode cache: %w", err)
	}
	return all, nil
}

func (s *Store) writeAll(all map[string]*CacheEntry) error {
	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}
