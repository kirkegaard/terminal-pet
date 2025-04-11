package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	// "strings"

	"github.com/kirkegaard/terminal-pet/pkg/config"
	"github.com/kirkegaard/terminal-pet/pkg/db"
	"github.com/kirkegaard/terminal-pet/pkg/db/repo"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	gossh "golang.org/x/crypto/ssh"

	"github.com/muesli/termenv"
)

var (
	hostKeyPath = ".ssh/id_ed25519"
)

type SSHServer struct {
	server        *ssh.Server
	petRepository *repo.PetRepository
	db            *db.DB
	serverCtx     context.Context
}

func NewSSHServer(ctx context.Context) (*SSHServer, error) {
	var err error

	cfg := config.FromContext(ctx)
	dbx := db.FromContext(ctx)
	if dbx == nil {
		return nil, fmt.Errorf("database not found in context")
	}

	petRepository := repo.NewPetRepository(dbx)

	s := &SSHServer{
		petRepository: petRepository,
		db:            dbx,
		serverCtx:     ctx,
	}

	hostKeyDir := filepath.Dir(hostKeyPath)
	if err := os.MkdirAll(hostKeyDir, 0700); err != nil {
		return nil, err
	}

	mw := []wish.Middleware{
		s.withDatabaseMiddleware(),
		WithPublicKeyMiddleware(),
		bm.MiddlewareWithProgramHandler(SessionHandler, termenv.TrueColor),
	}

	// mw = append(mw, func(h ssh.Handler) ssh.Handler {
	// 	return func(s ssh.Session) {
	// 		active := s.Pty()
	// 		if active {
	// 			environ := s.Environ()
	// 			var termVar string
	// 			for _, env := range environ {
	// 				if strings.HasPrefix(env, "TERM=") {
	// 					termVar = strings.TrimPrefix(env, "TERM=")
	// 					break
	// 				}
	// 			}
	// 		}
	// 		h(s)
	// 	}
	// })

	opt := []ssh.Option{
		wish.WithAddress(cfg.SSH.ListenAddr),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithPublicKeyAuth(s.publicKeyHandler),
		wish.WithMiddleware(mw...),
	}

	s.server, err = wish.NewServer(opt...)

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SSHServer) withDatabaseMiddleware() wish.Middleware {
	return func(handler ssh.Handler) ssh.Handler {
		return func(session ssh.Session) {
			sshCtx := session.Context()
			sshCtx.SetValue(db.ContextKeyDB, s.db)

			log.Debug("Adding database to context",
				"db_present", s.db != nil)

			handler(session)
		}
	}
}

func (s *SSHServer) publicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	pubKeyStr := fmt.Sprintf("%s %s", key.Type(), gossh.FingerprintSHA256(key))
	ctx.SetValue(string(PublicKeyKey), pubKeyStr)
	ctx.SetValue(db.ContextKeyDB, s.db)
	// log.Info("Public key auth", "user", ctx.User(), "key_fingerprint", gossh.FingerprintSHA256(key))
	return true
}

func (s *SSHServer) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *SSHServer) Serve(l net.Listener) error {
	return s.server.Serve(l)
}

func (s *SSHServer) Close() error {
	return s.server.Close()
}

func (s *SSHServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
