package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seraj/define/api"
	"github.com/seraj/define/audio"
	"github.com/seraj/define/dict"
	"github.com/seraj/define/stem"
)

var (
	wordStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	posStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	defStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	dimStyle  = lipgloss.NewStyle().Faint(true)
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	muted     = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8"))
	synStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	antStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
)

type lookupResultMsg struct {
	def *api.Definition
	err error
}

type audioState int

const (
	audioIdle audioState = iota
	audioPlaying
	audioDone
	audioError
)

type audioStateMsg struct {
	state audioState
	err   error
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

	player    *audio.Player
	audioSt   audioState
	audioErr  error
	audioCh   <-chan audioStateMsg
	cancel    context.CancelFunc
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

func (m *Model) SetPlayer(p *audio.Player) {
	m.player = p
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
		case "p":
			return m, m.startPlayback()
		case "s":
			if m.audioSt == audioPlaying && m.cancel != nil {
				m.cancel()
				m.audioSt = audioIdle
			}
			return m, nil
		}

	case lookupResultMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.def = msg.def
		}
		m.viewport.SetContent(m.renderTree())
		return m, nil

	case audioStateMsg:
		m.audioSt = msg.state
		m.audioErr = msg.err
		if msg.state == audioPlaying {
			return m, m.waitAudio()
		}
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
		phonetic = dimStyle.Render("  " + m.def.Phonetics[0].Text)
	}

	cached := ""
	if m.fromCache {
		cached = muted.Render(" (cached)")
	}

	audioBadge := ""
	switch m.audioSt {
	case audioPlaying:
		audioBadge = muted.Render("  ♫ playing...")
	case audioDone:
		audioBadge = muted.Render("  ♫ done")
	case audioError:
		audioBadge = errStyle.Render("  ♫ error")
	}

	playBadge := ""
	if m.hasAudio() && m.audioSt == audioIdle {
		playBadge = muted.Render("  ♫ play")
	}

	return wordStyle.PaddingLeft(2).Render(m.def.Word) + phonetic + cached + audioBadge + playBadge
}

func (m Model) renderTree() string {
	if m.err != nil {
		return errStyle.PaddingLeft(2).Render("✖  " + m.err.Error()) + "\n"
	}
	if m.def == nil {
		return ""
	}

	searchStem := stem.Stem(m.word)

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
			text := highlightInflections(d.Definition, searchStem, defStyle)

			if isLast {
				s += dimStyle.Render("      ├─ ") + num + text + "\n"
			} else {
				s += dimStyle.Render("  │   ├─ ") + num + text + "\n"
			}

			if d.Example != "" {
				exText := highlightInflections(d.Example, searchStem, dimStyle)

				if isLast {
					s += dimStyle.Render("      │    ╰─ ") + dimStyle.Render("\"") + exText + dimStyle.Render("\"") + "\n"
				} else {
					s += dimStyle.Render("  │   │    ╰─ ") + dimStyle.Render("\"") + exText + dimStyle.Render("\"") + "\n"
				}
			}
		}

		if !isLast {
			s += dimStyle.Render("  │") + "\n"
		}

		if len(meaning.Synonyms) > 0 {
			synLine := synStyle.Render("syn: ") + dimStyle.Render(strings.Join(meaning.Synonyms, " · "))
			if isLast {
				s += "      " + synLine + "\n"
			} else {
				s += "  │   " + synLine + "\n"
			}
		}
		if len(meaning.Antonyms) > 0 {
			antLine := antStyle.Render("ant: ") + dimStyle.Render(strings.Join(meaning.Antonyms, " · "))
			if isLast {
				s += "      " + antLine + "\n"
			} else {
				s += "  │   " + antLine + "\n"
			}
		}
	}

	return s
}

func highlightInflections(text, searchStem string, baseStyle lipgloss.Style) string {
	var result strings.Builder
	for _, word := range stem.Tokenize(text) {
		if word == "" {
			continue
		}
		s := stem.Stem(word)
		if s == searchStem {
			result.WriteString(wordStyle.Render(word))
		} else {
			result.WriteString(baseStyle.Render(word))
		}
	}
	return result.String()
}

func (m Model) renderFooter() string {
	parts := []string{
		muted.Render("[b] back"),
		muted.Render("[?] help"),
		muted.Render("[f] refresh"),
		muted.Render("[↑/↓] scroll"),
		muted.Render("[q] quit"),
	}

	if m.hasAudio() {
		if m.audioSt == audioPlaying {
			parts = append(parts, muted.Render("[s] stop"))
		} else if m.audioSt == audioError {
			parts = append(parts, muted.Render("[p] retry"))
		} else {
			parts = append(parts, muted.Render("[p] play"))
		}
	}

	if m.audioSt == audioError && m.audioErr != nil {
		parts = append(parts, errStyle.Render(m.audioErr.Error()))
	}

	return muted.Render(lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.JoinHorizontal(lipgloss.Top, parts...)))
}

func (m Model) hasAudio() bool {
	if m.player == nil || m.def == nil {
		return false
	}
	for _, p := range m.def.Phonetics {
		if p.Audio != "" {
			return true
		}
	}
	return false
}

func (m *Model) audioURL() string {
	if m.def == nil {
		return ""
	}
	for _, p := range m.def.Phonetics {
		if p.Audio != "" {
			return p.Audio
		}
	}
	return ""
}

func (m *Model) startPlayback() tea.Cmd {
	url := m.audioURL()
	if url == "" {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	ch := make(chan audioStateMsg, 1)
	m.audioCh = ch

	go func() {
		defer close(ch)
		ch <- audioStateMsg{state: audioPlaying}
		if err := m.player.Play(ctx, url); err != nil {
			ch <- audioStateMsg{state: audioError, err: err}
		} else {
			ch <- audioStateMsg{state: audioDone}
		}
	}()

	return m.waitAudio()
}

func (m *Model) waitAudio() tea.Cmd {
	ch := m.audioCh
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}
