package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	ti := textinput.New()
	ti.Placeholder = "Pikachu"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	m := model{
		width:     pty.Window.Width,
		height:    pty.Window.Height,
		choices:   []string{"Feed", "Toggle Lights", "Play", "Medicine", "Clean", "Stats", "Discipline", "Status"},
		selected:  make(map[int]struct{}),
		textInput: ti,
		keys:      keys,
		help:      help.New(),
		err:       nil,
	}

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

type (
	errMsg error
)

type model struct {
	width     int
	height    int
	cursor    int
	selected  map[int]struct{}
	choices   []string
	textInput textinput.Model
	keys      keyMap
	help      help.Model
	err       error
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.height = msg.Height
		m.width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Action):
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var output string

	// output = fmt.Sprintf("World time: %s\n", m.world.time.String())

	output += fmt.Sprintf(
		"What’s your pets name?\n\n%s",
		m.textInput.View(),
	) + "\n"

	helpView := m.help.View(m.keys)
	height := 8 - strings.Count(output, "\n") - strings.Count(helpView, "\n")

	return "\n" + output + strings.Repeat("\n", height) + helpView
}
