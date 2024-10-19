package ssh

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type timeMsg time.Time

type UI struct {
	Renderer *lipgloss.Renderer
	time     time.Time
	width    int
	height   int
}

func NewUI(ctx context.Context, renderer *lipgloss.Renderer, width int, height int) *UI {
	ui := &UI{
		Renderer: renderer,
		width:    width,
		height:   height,
		time:     time.Now(),
	}

	return ui
}

// Init implements tea.Model.
func (ui *UI) Init() tea.Cmd {
	return nil
}

func (ui *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timeMsg:
		ui.time = time.Time(msg)
	case tea.WindowSizeMsg:
		ui.height = msg.Height
		ui.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return ui, tea.Quit
		}
	}
	return ui, nil
}

func (ui *UI) View() string {
	s := "Your window size is x: %d y: %d\n"
	s += "Time: " + ui.time.Format(time.RFC1123) + "\n\n"
	s += "Press 'q' to quit\n"
	return fmt.Sprintf(s, ui.width, ui.height)
}
