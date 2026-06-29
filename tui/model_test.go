package tui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seraj/define/api"
	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
)

func TestModelViewShowsDefinition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"word":"hello","phonetics":[{"text":"/həˈloʊ/"}],"meanings":[{"partOfSpeech":"interjection","definitions":[{"definition":"A greeting.","example":"Hello, world!"}]}]}]`))
	}))
	defer srv.Close()

	client := api.NewClient(api.WithBaseURL(srv.URL))

	store, err := cache.NewStore(t.TempDir() + "/cache.json")
	if err != nil {
		t.Fatal(err)
	}

	svc := dict.NewService(client, store)
	m := NewModel("hello", svc)

	cmd := m.Init()
	msg := cmd()
	next, _ := m.Update(msg)

	view := next.View()
	if !strings.Contains(view, "hello") {
		t.Error("view should contain the word")
	}
	if !strings.Contains(view, "interjection") {
		t.Error("view should contain POS label")
	}
	if !strings.Contains(view, "A greeting.") {
		t.Error("view should contain definition")
	}
}

func TestModelKeyQuit(t *testing.T) {
	svc := dict.NewService(nil, nil)
	m := NewModel("test", svc)
	m.def = &api.Definition{Word: "test"}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	nextM, ok := next.(Model)
	if !ok {
		t.Error("ctrl+c should return a Model")
		return
	}
	if !nextM.quitting {
		t.Error("ctrl+c should set quitting flag")
	}
}

func TestModelViewShowsFooter(t *testing.T) {
	svc := dict.NewService(nil, nil)
	m := NewModel("test", svc)
	m.def = &api.Definition{Word: "test"}

	view := m.View()
	if !strings.Contains(view, "[q] quit") {
		t.Error("view should contain footer with '[q] quit'")
	}
}

func TestModelViewShowsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()

	client := api.NewClient(api.WithBaseURL(srv.URL))
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	svc := dict.NewService(client, store)
	m := NewModel("xyzzy", svc)

	cmd := m.Init()
	msg := cmd()
	next, _ := m.Update(msg)

	view := next.View()
	if !strings.Contains(view, "not found") {
		t.Error("view should show error for not-found word")
	}
	if !strings.Contains(view, "xyzzy") {
		t.Error("view should contain the queried word in error display")
	}
}

func TestModelForceRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"word":"hello","meanings":[{"partOfSpeech":"noun","definitions":[{"definition":"test"}]}]}]`))
	}))
	defer srv.Close()

	client := api.NewClient(api.WithBaseURL(srv.URL))
	store, _ := cache.NewStore(t.TempDir() + "/cache.json")
	svc := dict.NewService(client, store)
	m := NewModel("hello", svc)
	m.def = &api.Definition{Word: "hello"}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if cmd == nil {
		t.Error("'f' key should return a lookup command")
		return
	}

	msg := cmd()
	result, ok := msg.(lookupResultMsg)
	if !ok {
		t.Error("cmd should return a lookupResultMsg")
		return
	}
	if result.err != nil {
		t.Errorf("unexpected lookup error: %v", result.err)
	}
	if next.View() == "" {
		t.Error("model should still render")
	}
}
