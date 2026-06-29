package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
)

type screen int

const (
	screenHistory screen = iota
	screenDefinition
)

type AppModel struct {
	screen   screen
	history  HistoryModel
	defModel Model
}

func NewAppModel(word string, svc *dict.Service, store *cache.Store) AppModel {
	a := AppModel{
		history: NewHistoryModel(svc, store),
	}

	if word == "" {
		a.screen = screenHistory
	} else {
		a.screen = screenDefinition
		a.defModel = NewModel(word, svc)
	}

	return a
}

func (a AppModel) Init() tea.Cmd {
	if a.screen == screenDefinition {
		return a.defModel.Init()
	}
	return a.history.Init()
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case lookupResultMsg:
		if msg.err != nil {
			a.screen = screenHistory
			a.history.reloadItems()
			return a, nil
		}
		if a.screen == screenDefinition {
			a.defModel.def = msg.def
			a.defModel.fromCache = false
			a.defModel.viewport.SetContent(a.defModel.renderTree())
			return a, nil
		}
		a.defModel = NewModel(msg.def.Word, a.history.service)
		a.defModel.def = msg.def
		a.defModel.fromCache = true
		a.defModel.viewport.SetContent(a.defModel.renderTree())
		a.screen = screenDefinition
		return a, tea.WindowSize()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q", "esc":
			if a.screen == screenDefinition {
				a.screen = screenHistory
				a.history.reloadItems()
				return a, nil
			}
			return a, tea.Quit
		case "b":
			if a.screen == screenDefinition {
				a.screen = screenHistory
				a.history.reloadItems()
				return a, nil
			}
		}

	case tea.WindowSizeMsg:
		if a.screen == screenHistory {
			a.history.Update(msg)
		} else {
			a.defModel.Update(msg)
		}
	}

	if a.screen == screenHistory {
		next, cmd := a.history.Update(msg)
		a.history = next.(HistoryModel)
		if a.history.quitting {
			return a, tea.Quit
		}
		return a, cmd
	}

	next, cmd := a.defModel.Update(msg)
	a.defModel = next.(Model)
	return a, cmd
}

func (a AppModel) View() string {
	if a.screen == screenHistory {
		return a.history.View()
	}
	return a.defModel.View()
}
