package ssh

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	bm "github.com/charmbracelet/wish/bubbletea"
)

// Our bubble tea session handler
func SessionHandler(s ssh.Session) *tea.Program {
	pty, _, active := s.Pty()
	if !active {
		return nil
	}

	ctx := s.Context()

	renderer := bm.MakeRenderer(s)

	log.Info("Creating UI")
	m := NewUI(ctx, renderer, pty.Window.Width, pty.Window.Height)

	opts := bm.MakeOptions(s)
	opts = append(opts,
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)

	log.Info("Creating program")
	p := tea.NewProgram(m, opts...)

	log.Info("Returning program")
	return p
}
