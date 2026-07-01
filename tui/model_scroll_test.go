package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seraj/define/api"
	"github.com/seraj/define/dict"
)

func TestModelScrollDown(t *testing.T) {
	defs := make([]api.Def, 50)
	for i := range defs {
		defs[i] = api.Def{Definition: strings.Repeat("x", 70)}
	}

	m := NewModel("test", dict.NewService(nil, nil))
	m.def = &api.Definition{
		Word: "test",
		Meanings: []api.Meaning{
			{PartOfSpeech: "noun", Definitions: defs},
		},
	}
	m.viewport.Height = 10
	m.viewport.Width = 70
	m.viewport.SetContent(m.renderTree())

	if !m.viewport.AtTop() {
		t.Error("viewport should be at top initially")
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if next.(Model).viewport.YOffset <= 0 {
		t.Error("KeyDown should scroll (YOffset > 0)")
	}

	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if next.(Model).viewport.YOffset <= 1 {
		t.Error("'j' should scroll further")
	}
}
