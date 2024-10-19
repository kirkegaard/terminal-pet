package ssh

import (
	"context"
	"net"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
)

var users = map[string]string{}

type SSHServer struct {
	server *ssh.Server
}

// Server is the SSH Server
func NewSSHServer(host string, port string) (*SSHServer, error) {
	var err error
	s := &SSHServer{}

	opt := []ssh.Option{
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			// @TODO Look up the user in the database. If the user is not found,
			// create a new user.
			for _, pubkey := range users {
				parsed, _, _, _, _ := ssh.ParseAuthorizedKey(
					[]byte(pubkey),
				)

				// User is found
				if ssh.KeysEqual(key, parsed) {
					log.Info("User found", "user", ctx.User())
					return true
				} else {
					log.Info("User not found", "user", ctx.User())
					// Create user
					return false
				}
			}
			// Worst case we didnt find the user or could not create a new user
			return false
		}),
		wish.WithMiddleware(
			logging.Middleware(),
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					log.Info("New session", "user", sess.User(), "remote", sess.RemoteAddr())
				}
			},
		),
	}

	s.server, err = wish.NewServer(opt...)

	if err != nil {
		return nil, err
	}

	return s, nil
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
