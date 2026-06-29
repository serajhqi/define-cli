package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seraj/define/api"
	"github.com/seraj/define/dict"
)

var (
	wordStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	posStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	defStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	dimStyle  = lipgloss.NewStyle().Faint(true)
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	muted     = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8"))
)

type lookupResultMsg struct {
	def *api.Definition
	err error
}

type Model struct {
	word      string
	def       *api.Definition
	err       error
	service   *dict.Service
	viewport  viewport.Model
	quitting  bool
	fromCache bool
	width     int
}

func NewModel(word string, service *dict.Service) Model {
	vp := viewport.New(70, 20)
	vp.KeyMap = viewport.KeyMap{}

	return Model{
		word:     word,
		service:  service,
		viewport: vp,
		width:    70,
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		def, err := m.service.LookupDefinition(m.word, false)
		return lookupResultMsg{def: def, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "f":
			return m, func() tea.Msg {
				def, err := m.service.LookupDefinition(m.word, true)
				return lookupResultMsg{def: def, err: err}
			}
		case "up", "down", "j", "k":
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case lookupResultMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.def = msg.def
		}
		m.viewport.SetContent(m.renderTree())
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	header := m.renderHeader()
	body := m.viewport.View()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		body,
		footer,
	)
}

func (m Model) renderHeader() string {
	if m.def == nil {
		return wordStyle.PaddingLeft(2).Render(m.word)
	}

	phonetic := ""
	if len(m.def.Phonetics) > 0 {
		phonetic = dimStyle.Render("  " + m.def.Phonetics[0])
	}

	cached := ""
	if m.fromCache {
		cached = muted.Render(" (cached)")
	}

	return wordStyle.PaddingLeft(2).Render(m.def.Word) + phonetic + cached
}

func (m Model) renderTree() string {
	if m.err != nil {
		return errStyle.PaddingLeft(2).Render("✖  " + m.err.Error()) + "\n"
	}
	if m.def == nil {
		return ""
	}

	var s string
	for mi, meaning := range m.def.Meanings {
		isLast := mi == len(m.def.Meanings)-1

		if isLast {
			s += dimStyle.Render("  ╰─▸ ") + posStyle.Render(meaning.PartOfSpeech) + "\n"
		} else {
			s += dimStyle.Render("  ├─▸ ") + posStyle.Render(meaning.PartOfSpeech) + "\n"
		}

		for di, d := range meaning.Definitions {
			num := fmt.Sprintf("%d. ", di+1)

			if isLast {
				s += dimStyle.Render("      ├─ ") + defStyle.Render(num+d.Definition) + "\n"
			} else {
				s += dimStyle.Render("  │   ├─ ") + defStyle.Render(num+d.Definition) + "\n"
			}

			if d.Example != "" {
				if isLast {
					s += dimStyle.Render("      │    ╰─ ") + dimStyle.Render("\""+d.Example+"\"") + "\n"
				} else {
					s += dimStyle.Render("  │   │    ╰─ ") + dimStyle.Render("\""+d.Example+"\"") + "\n"
				}
			}
		}

		if !isLast {
			s += dimStyle.Render("  │") + "\n"
		}
	}

	return s
}

func (m Model) renderFooter() string {
	parts := []string{
		muted.Render("[b] back"),
		muted.Render("[?] help"),
		muted.Render("[f] refresh"),
		muted.Render("[↑/↓] scroll"),
		muted.Render("[q] quit"),
	}
	return muted.Render(lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.JoinHorizontal(lipgloss.Top, parts...)))
}
