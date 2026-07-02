package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seraj/define/cache"
	"github.com/seraj/define/dict"
)

type HistoryModel struct {
	service    *dict.Service
	store      *cache.Store
	input      textinput.Model
	items      []cache.HistoryItem
	filtered   []cache.HistoryItem
	cursor     int
	confirming bool
	confirmIdx int
	showLookup bool
	width      int
	height     int
	quitting   bool
}

func NewHistoryModel(service *dict.Service, store *cache.Store) *HistoryModel {
	ti := textinput.New()
	ti.Placeholder = "Filter..."
	ti.Focus()
	ti.CharLimit = 50

	items, _ := store.List()

	return &HistoryModel{
		service:  service,
		store:    store,
		input:    ti,
		items:    items,
		filtered: items,
		width:    70,
	}
}

func (m *HistoryModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.confirming {
		return m.updateConfirm(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down":
			max := len(m.filtered) - 1
			if m.showLookup && len(m.filtered) == 0 {
				max = 0
			}
			if m.showLookup {
				max++
			}
			if m.cursor < max {
				m.cursor++
			}
			return m, nil

		case "enter":
			if m.showLookup && m.cursor >= len(m.filtered) {
				return m, func() tea.Msg {
					def, err := m.service.LookupDefinition(m.input.Value(), false)
					return lookupResultMsg{def: def, err: err}
				}
			}
			if m.cursor >= 0 && m.cursor < len(m.filtered) {
				word := m.filtered[m.cursor].Word
				return m, func() tea.Msg {
					def, err := m.service.LookupDefinition(word, false)
					return lookupResultMsg{def: def, err: err}
				}
			}
			return m, nil

		case "delete":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.confirming = true
				m.confirmIdx = m.cursor
			}
			return m, nil

		case "esc":
			m.input.SetValue("")
			m.filterItems()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.filterItems()
	return m, cmd
}

func (m *HistoryModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "y":
		word := m.filtered[m.confirmIdx].Word
		m.store.Delete(word)
		items, _ := m.store.List()
		m.items = items
		m.filterItems()
		m.confirming = false
		if m.cursor >= len(m.filtered) && m.cursor > 0 {
			m.cursor = len(m.filtered) - 1
		}
		return m, nil
	case "n", "esc":
		m.confirming = false
		return m, nil
	}
	return m, nil
}

func (m *HistoryModel) reloadItems() {
	items, _ := m.store.List()
	m.items = items
	m.filterItems()
}

func (m *HistoryModel) filterItems() {
	filter := m.input.Value()
	m.filtered = nil
	m.showLookup = false

	for _, item := range m.items {
		if fuzzyMatch(filter, item.Word) {
			m.filtered = append(m.filtered, item)
		}
	}

	if filter != "" && len(m.filtered) == 0 {
		m.showLookup = true
	}

	if m.cursor >= len(m.filtered) {
		if m.showLookup {
			m.cursor = len(m.filtered)
		} else {
			m.cursor = 0
		}
	}
}

func (m *HistoryModel) View() string {
	if m.quitting {
		return ""
	}

	if m.confirming {
		return m.renderConfirm()
	}

	start := 0
	end := len(m.filtered)

	if m.height > 0 {
		headerRows := 2
		footerRows := 1
		lookupRows := 0
		if m.showLookup {
			lookupRows = 2
		}

		availableRows := m.height - headerRows - footerRows - lookupRows
		if availableRows < 0 {
			availableRows = 0
		}

		itemsPerRow := 3
		maxVisible := availableRows / itemsPerRow
		start, end = m.visibleRange(maxVisible)
	}

	var b strings.Builder

	b.WriteString(muted.Render(lipgloss.NewStyle().PaddingLeft(2).Render("Filter: ")))
	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	if len(m.filtered) == 0 && !m.showLookup {
		b.WriteString(muted.Render("  No cached words yet.\n"))
	}

	for i := start; i < end; i++ {
		item := m.filtered[i]
		cursor := "  "
		if m.cursor == i {
			cursor = bold.Render("> ")
		}

		preview := item.Preview
		maxPreview := m.width - 7
		if len(preview) > maxPreview {
			preview = preview[:maxPreview-3] + "..."
		}

		wordStr := wordStyle.Render(item.Word)
		dateStr := muted.Render(item.LookedUpAt.Format("2006-01-02 15:04"))
		padding := m.width - 6 - lipgloss.Width(wordStr) - lipgloss.Width(dateStr)
		if padding < 1 {
			padding = 1
		}

		b.WriteString(fmt.Sprintf("%s%s%s%s\n", cursor, wordStr, strings.Repeat(" ", padding), dateStr))
		b.WriteString(fmt.Sprintf("     %s\n", defStyle.Render(preview)))

		if i < len(m.filtered)-1 || m.showLookup {
			dots := (m.width - 4) / 3
			if dots > 30 {
				dots = 30
			}
			b.WriteString("  " + muted.Render(strings.Repeat("· ", dots)) + "\n")
		}
	}

	if m.showLookup {
		cursor := "  "
		if m.cursor >= len(m.filtered) {
			cursor = bold.Render("> ")
		}
		b.WriteString(fmt.Sprintf("%s%s '%s'  [Enter]\n", cursor, wordStyle.Render("Look up →"), wordStyle.Render(m.input.Value())))
		b.WriteString("\n")
	}

	footerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("15")).
		Foreground(lipgloss.Color("0")).
		Padding(0, 2).
		Width(m.width)
	footer := footerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top,
		"[↑/↓] nav",
		"[enter] select",
		"[del] delete",
		"[q] quit",
	))
	b.WriteString(footer)

	return b.String()
}

func (m *HistoryModel) visibleRange(maxVisible int) (int, int) {
	total := len(m.filtered)
	if total <= maxVisible {
		return 0, total
	}

	half := maxVisible / 2
	start := m.cursor - half
	end := start + maxVisible

	if start < 0 {
		start = 0
		end = maxVisible
	}
	if end > total {
		end = total
		start = total - maxVisible
	}

	return start, end
}

func (m *HistoryModel) renderConfirm() string {
	if m.confirmIdx < 0 || m.confirmIdx >= len(m.filtered) {
		m.confirming = false
		return m.View()
	}

	word := m.filtered[m.confirmIdx].Word
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  Delete \"%s\" from cache?\n\n", wordStyle.Render(word)))
	b.WriteString(fmt.Sprintf("  %s  %s\n", muted.Render("[y] yes"), muted.Render("[n] no")))
	return b.String()
}

var bold = lipgloss.NewStyle().Bold(true)
