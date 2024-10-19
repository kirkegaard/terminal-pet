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
	logger    *log.Logger
	ctx       context.Context
}

func NewServer() (*Server, error) {
	var err error

	// @TODO Figure out how we can use context to set configuration here
	s := &Server{}

	// Add database
	// @TODO Get configuration from shared context
	s.DB, err = db.Open(dbDriverName, dbDSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Add world
	s.World = world.NewWorld(time.Now())

	// Add ssh server
	// @TODO Get configuration from shared context
	s.SSHServer, err = ssh.NewSSHServer(sshHost, sshPort)
	if err != nil {
		return nil, fmt.Errorf("create ssh server: %w", err)
	}

	return s, nil
}

func main() {
	s, err := NewServer()
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
