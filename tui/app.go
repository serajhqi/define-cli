package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seraj/define/audio"
	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
)

type screen int

const (
	screenHistory screen = iota
	screenDefinition
	screenHelp
)

type AppModel struct {
	screen   screen
	history  HistoryModel
	defModel Model
	player   *audio.Player
}

func NewAppModel(word string, svc *dict.Service, store *cache.Store, player *audio.Player) AppModel {
	a := AppModel{
		history: NewHistoryModel(svc, store),
		player:  player,
	}

	if word == "" {
		a.screen = screenHistory
	} else {
		a.screen = screenDefinition
		a.defModel = NewModel(word, svc)
		a.defModel.SetPlayer(player)
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
		a.defModel.SetPlayer(a.player)
		a.defModel.def = msg.def
		a.defModel.fromCache = true
		a.defModel.viewport.SetContent(a.defModel.renderTree())
		a.screen = screenDefinition
		return a, tea.WindowSize()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "?":
			if a.screen != screenHelp {
				a.screen = screenHelp
				return a, nil
			}
		case "q", "esc":
			if a.screen == screenHelp {
				a.screen = screenHistory
				return a, nil
			}
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
	if a.screen == screenHelp {
		return a.renderHelp()
	}
	if a.screen == screenHistory {
		return a.history.View()
	}
	return a.defModel.View()
}

func (a AppModel) renderHelp() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).PaddingLeft(2).Render("define — Dictionary CLI\n\n")
	sections := []struct {
		header string
		items  []string
	}{
		{"Keybindings",
			[]string{
				"[p]      play pronunciation",
				"[s]      stop playback",
				"[f]      force refresh (bypass cache)",
				"[b]      back to history",
				"[?]      this help screen",
				"[↑/↓]    scroll / navigate",
				"[q]      quit / back",
				"[esc]    quit / back",
				"[ctrl+c]  quit",
			},
		},
		{"History Browser",
			[]string{
				"[↑/↓]    navigate list",
				"[enter]   select word / look up",
				"[d]       delete cached word",
			},
		},
		{"CLI Flags",
			[]string{
				"--history  launch TUI history browser",
				"--play     auto-play pronunciation",
				"--plain    plain text output, no TUI",
				"-f         force refresh API call",
				"--help     show this help",
			},
		},
	}

	var s string
	s += title
	for _, sec := range sections {
		s += muted.Render("  "+sec.header) + "\n"
		for _, item := range sec.items {
			s += muted.Render("    " + item) + "\n"
		}
		s += "\n"
	}
	s += muted.Render("  Press [q] or [esc] to go back.\n")
	return s
}
