package ssh

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	bm "github.com/charmbracelet/wish/bubbletea"
	"time"
)

// Our bubble tea session handler
func SessionHandler(s ssh.Session) *tea.Program {
	pty, _, active := s.Pty()
	if !active {
		return nil
	}

	ctx := s.Context()

	renderer := bm.MakeRenderer(s)

	m := NewUI(ctx, renderer, pty.Window.Width, pty.Window.Height)

	opts := bm.MakeOptions(s)
	opts = append(opts,
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)

	p := tea.NewProgram(m, opts...)

	go func() {
		for {
			<-time.After(1 * time.Second)
			p.Send(timeMsg(time.Now()))
		}
	}()

	return p
}
