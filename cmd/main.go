package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	cssh "github.com/charmbracelet/ssh"

	"github.com/kirkegaard/terminal-pet/pkg/config"
	"github.com/kirkegaard/terminal-pet/pkg/db"
	"github.com/kirkegaard/terminal-pet/pkg/ssh"
	"github.com/kirkegaard/terminal-pet/pkg/world"
)

var (
	sshHost      = "localhost"
	sshPort      = "23234"
	dbDriverName = "sqlite3"
	dbDSN        = "file::memory:?cache=shared"
)

type Server struct {
	SSHServer *ssh.SSHServer
	World     *world.World
	DB        *db.DB
	Config    *config.Config
	logger    *log.Logger
	ctx       context.Context
}

func NewServer(ctx context.Context) (*Server, error) {
	var err error

	cfg := config.FromContext(ctx)

	// Open database connection
	dbx, err := db.Open(ctx, cfg.DB.Driver, cfg.DB.DataSource)
	if err != nil {
		log.Info("Could not open database", "error", err)
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Add db to context
	ctx = db.WithContext(ctx, dbx)

	s := &Server{
		Config: cfg,
		DB:     dbx,
		ctx:    ctx,
	}

	// Add world
	s.World = world.NewWorld(time.Now())

	// Add ssh server
	s.SSHServer, err = ssh.NewSSHServer(ctx)
	if err != nil {
		return nil, fmt.Errorf("create ssh server: %w", err)
	}

	return s, nil
}

func main() {
	ctx := context.Background()
	cfg := config.DefaultConfig()

	// @TODO Add a way to load config from file. Basically like this
	// if err := cfg.ParseEnv(); err != nil {
	//  log.Fatal("Could not parse env", "error", err)
	// }

	// Set the config in the context
	ctx = config.WithContext(ctx, cfg)

	s, err := NewServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// @TODO Add more servers like api, websockets, admin, etc.
	// @TODO Look into errgroup for handling multiple servers
	// @TODO Add a way to gracefully shutdown all servers

	log.Info("Starting SSH server", "host", sshHost, "port", sshPort)
	go func() {
		if err = s.SSHServer.ListenAndServe(); err != nil && !errors.Is(err, cssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done

	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.SSHServer.Shutdown(ctx); err != nil && !errors.Is(err, cssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}
