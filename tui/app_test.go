package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
)

func TestAppStartsWithHistory(t *testing.T) {
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	svc := dict.NewService(nil, store)
	app := NewAppModel("", svc, store)

	if app.screen != screenHistory {
		t.Error("app with no word should start in history screen")
	}
	if app.View() == "" {
		t.Error("history view should not be empty")
	}
}

func TestAppStartsWithDefinition(t *testing.T) {
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	svc := dict.NewService(nil, store)
	app := NewAppModel("hello", svc, store)

	if app.screen != screenDefinition {
		t.Error("app with word should start in definition screen")
	}
}

func TestAppQuitsOnCtrlC(t *testing.T) {
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	svc := dict.NewService(nil, store)
	app := NewAppModel("", svc, store)

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should return a quit command")
	}
}

func TestAppQGoesBackFromDefinition(t *testing.T) {
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	svc := dict.NewService(nil, store)
	app := NewAppModel("hello", svc, store)
	app.defModel.def = nil // simulate initial lookup error or no data

	next, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	nextApp, ok := next.(AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if nextApp.screen != screenHistory {
		t.Error("'q' in definition should go back to history")
	}
}
