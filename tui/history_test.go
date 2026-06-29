package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
)

func TestHistoryLoadsItems(t *testing.T) {
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	store.Set("hello", &cache.CacheEntry{
		Preview:    "A greeting.",
		LookedUpAt: time.Now(),
	})
	store.Set("world", &cache.CacheEntry{
		Preview:    "The Earth.",
		LookedUpAt: time.Now().Add(-time.Hour),
	})

	svc := dict.NewService(nil, store)
	m := NewHistoryModel(svc, store)

	view := m.View()
	if !strings.Contains(view, "hello") {
		t.Error("view should show cached word 'hello'")
	}
	if !strings.Contains(view, "world") {
		t.Error("view should show cached word 'world'")
	}
	if !strings.Contains(view, "A greeting.") {
		t.Error("view should show preview text")
	}
}

func TestHistoryFuzzyFilter(t *testing.T) {
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	store.Set("hello", &cache.CacheEntry{Preview: "...", LookedUpAt: time.Now()})
	store.Set("help", &cache.CacheEntry{Preview: "...", LookedUpAt: time.Now()})
	store.Set("world", &cache.CacheEntry{Preview: "...", LookedUpAt: time.Now()})

	svc := dict.NewService(nil, store)
	m := NewHistoryModel(svc, store)

	m.input.SetValue("hel")
	m.filterItems()

	view := m.View()
	if !strings.Contains(view, "hello") {
		t.Error("filter 'hel' should match 'hello'")
	}
	if !strings.Contains(view, "help") {
		t.Error("filter 'hel' should match 'help'")
	}
	if strings.Contains(view, "world") {
		t.Error("filter 'hel' should NOT match 'world'")
	}
}

func TestHistoryLookupItem(t *testing.T) {
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	svc := dict.NewService(nil, store)
	m := NewHistoryModel(svc, store)

	m.input.SetValue("xyz")
	m.filterItems()

	view := m.View()
	if !strings.Contains(view, "Look up") {
		t.Error("should show 'Look up' item when no match and filter non-empty")
	}
	if !strings.Contains(view, "xyz") {
		t.Error("'Look up' item should include the filter text")
	}
}

func TestHistoryDeleteConfirm(t *testing.T) {
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	store.Set("hello", &cache.CacheEntry{Preview: "...", LookedUpAt: time.Now()})

	svc := dict.NewService(nil, store)
	m := NewHistoryModel(svc, store)

	m.confirmIdx = 0
	m.confirming = true

	view := m.View()
	if !strings.Contains(view, "[y]") || !strings.Contains(view, "[n]") {
		t.Error("confirm dialog should show [y] yes [n] no")
	}
}

func TestHistoryCursorDownSelects(t *testing.T) {
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	store.Set("a", &cache.CacheEntry{Preview: "...", LookedUpAt: time.Now()})
	store.Set("b", &cache.CacheEntry{Preview: "...", LookedUpAt: time.Now()})

	svc := dict.NewService(nil, store)
	m := NewHistoryModel(svc, store)

	if m.cursor != 0 {
		t.Error("cursor should start at 0")
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	nextM, ok := next.(HistoryModel)
	if !ok {
		t.Fatal("expected HistoryModel")
	}
	if nextM.cursor != 1 {
		t.Error("cursor should move down")
	}
}
