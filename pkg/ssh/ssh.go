package ssh

import (
	"context"
	"net"

	"github.com/kirkegaard/terminal-pet/pkg/config"
	"github.com/kirkegaard/terminal-pet/pkg/db"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	// "github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
)

var users = map[string]string{}

type SSHServer struct {
	server *ssh.Server
}

// Server is the SSH Server
func NewSSHServer(ctx context.Context) (*SSHServer, error) {
	var err error

	cfg := config.FromContext(ctx)
	dbx := db.FromContext(ctx)

	log.Info("Database", "dbx", dbx)

	s := &SSHServer{}

	mw := []wish.Middleware{
		// BubbleTea middleware
		bm.MiddlewareWithProgramHandler(SessionHandler, termenv.ANSI256),
	}

	opt := []ssh.Option{
		wish.WithAddress(cfg.SSH.ListenAddr),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(PublicKeyHandler),
		wish.WithMiddleware(mw...),
	}

	s.server, err = wish.NewServer(opt...)

	if err != nil {
		return nil, err
	}

	return s, nil
}

func PublicKeyHandler(ctx ssh.Context, pk ssh.PublicKey) (allowed bool) {
	return true
}

// ListenAndServe starts the SSH server.
func (s *SSHServer) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Serve starts the SSH server on the given net.Listener.
func (s *SSHServer) Serve(l net.Listener) error {
	return s.server.Serve(l)
}

// Close closes the SSH server.
func (s *SSHServer) Close() error {
	return s.server.Close()
}

// Shutdown gracefully shuts down the SSH server.
func (s *SSHServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
