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
)

type Server struct {
	SSHServer *ssh.SSHServer
	DB        *db.DB
	Config    *config.Config
	logger    *log.Logger
	ctx       context.Context
}

func NewServer(ctx context.Context) (*Server, error) {
	var err error

	cfg := config.FromContext(ctx)
	if cfg == nil {
		log.Fatal("Config not found in context")
	}

	log.Info("Setting up database", "driver", cfg.DB.Driver, "data_source", cfg.DB.DataSource)

	// Create directory for database
	dbDir := "./tmp"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	// Open database connection
	dbx, err := db.Open(ctx, cfg.DB.Driver, cfg.DB.DataSource)
	if err != nil {
		log.Error("Could not open database", "error", err)
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Verify database connection
	if err := dbx.Ping(); err != nil {
		log.Error("Database ping failed", "error", err)
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Initialize database schema
	if err := dbx.CreateTables(ctx); err != nil {
		log.Error("Could not create database tables", "error", err)
		return nil, fmt.Errorf("create database tables: %w", err)
	}

	// Create a new context with the database
	dbCtx := db.WithContext(ctx, dbx)

	s := &Server{
		Config: cfg,
		DB:     dbx,
		ctx:    dbCtx,
	}

	s.SSHServer, err = ssh.NewSSHServer(dbCtx)
	if err != nil {
		return nil, fmt.Errorf("create ssh server: %w", err)
	}

	return s, nil
}

func main() {
	// Set up logging
	log.SetLevel(log.DebugLevel)

	// Create base context
	ctx := context.Background()

	// Load configuration with defaults
	cfg := config.DefaultConfig()

	// Parse environment variables
	if err := config.ParseEnv(cfg); err != nil {
		log.Warn("Failed to parse environment variables", "error", err)
	}

	log.Info("Configuration loaded",
		"ssh_listen", cfg.SSH.ListenAddr,
		"ssh_url", cfg.SSH.PublicURL,
		"db_driver", cfg.DB.Driver,
		"db_source", cfg.DB.DataSource)

	// Set the config in the context
	ctx = config.WithContext(ctx, cfg)

	// Create server
	s, err := NewServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// @TODO Add a way to gracefully shutdown all servers
	log.Info("Starting SSH server", "address", s.Config.SSH.ListenAddr)

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
