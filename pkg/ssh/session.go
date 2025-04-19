package ssh

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/kirkegaard/terminal-pet/pkg/db"
	"github.com/kirkegaard/terminal-pet/pkg/db/repo"
	"github.com/kirkegaard/terminal-pet/pkg/pet"
	petui "github.com/kirkegaard/terminal-pet/pkg/ui"
)

type PetSaveMsg struct {
	Pet       *pet.Pet
	PublicKey string
}

func SessionHandler(s ssh.Session) *tea.Program {
	pty, _, active := s.Pty()
	if !active {
		return nil
	}

	sessionCtx := s.Context()

	log.Debug("Session context keys", "keys", getContextKeys(sessionCtx))

	log.Debug("Session context type", "type", fmt.Sprintf("%T", sessionCtx))

	var dbx *db.DB = db.GetInstance()
	if dbx == nil {
		log.Error("Global database instance not available")
		fmt.Fprintln(s, "Error: Database connection is not available")
		return nil
	}

	log.Info("Using global database instance")

	petRepo := repo.NewPetRepository(dbx)
	log.Debug("Created pet repository", "repo", petRepo != nil)

	publicKey, ok := sessionCtx.Value(string(PublicKeyKey)).(string)
	if !ok || publicKey == "" {
		log.Warn("No public key found in session context", "user", s.User())
		publicKey = fmt.Sprintf("user-%s", s.User()) // Fallback for password auth
	}

	log.Debug("Using public key", "key", publicKey)

	existingPet, err := petRepo.FindByParentPublicKey(context.Background(), publicKey)
	if err != nil {
		log.Error("Error finding pet", "error", err)
	}

	renderer := bm.MakeRenderer(s)

	var ui *UI
	if existingPet != nil {
		log.Info("Found existing pet", "name", existingPet.Name, "user", s.User())

		log.Info("Pet health status",
			"health", existingPet.Health,
			"is_dead", existingPet.IsDead(),
			"is_dead_check", existingPet.Health <= 0)

		timePassed := time.Since(existingPet.LastVisit)
		log.Info("Time since last visit", "duration", timePassed.String())

		if !existingPet.IsDead() {
			simulateTimePassed(existingPet, timePassed)
		}

		ui = NewUI(context.Background(), renderer, pty.Window.Width, pty.Window.Height, existingPet, publicKey)

		if existingPet.Health <= 0 {
			log.Info("Pet is dead on connection, showing game over screen", "name", existingPet.Name)

			if petUIModel, ok := ui.petUI.(*petui.PetUI); ok {
				petUIModel.SetGameOver(true)
			} else {
				petUIValue := reflect.ValueOf(ui.petUI)
				setGameOverMethod := petUIValue.MethodByName("SetGameOver")
				if setGameOverMethod.IsValid() {
					setGameOverMethod.Call([]reflect.Value{reflect.ValueOf(true)})
				} else {
					log.Error("Could not find SetGameOver method")
				}
			}
		}
	} else {
		log.Info("Creating new pet for user", "user", s.User())

		parent := pet.NewParent(0, s.User())
		newPet := pet.NewPet(fmt.Sprintf("%s's pet", s.User()), time.Now(), parent)
		newPet.Happiness = 80
		newPet.Health = 100

		ui = NewUI(context.Background(), renderer, pty.Window.Width, pty.Window.Height, newPet, publicKey)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create a WaitGroup to ensure all goroutines are properly cleaned up
	var wg sync.WaitGroup

	opts := bm.MakeOptions(s)
	opts = append(opts,
		tea.WithAltScreen(),
		tea.WithContext(ctx),
		tea.WithoutCatchPanics(), // Let panics propagate for better error reporting
	)

	p := tea.NewProgram(ui, opts...)

	// Add a finalizer to handle shutdown cleanly
	shutdownOnce := &sync.Once{}
	shutdown := func() {
		shutdownOnce.Do(func() {
			log.Debug("Running clean shutdown")

			// Cancel the context first to signal all goroutines to stop
			cancel()

			// Wait for goroutines to finish
			wg.Wait()

			// Final state save
			if ui.currentPet != nil {
				ui.syncPetState()
				err := petRepo.Save(context.Background(), ui.currentPet, ui.publicKey)
				if err != nil {
					log.Error("Error saving final pet state", "error", err)
				} else {
					log.Info("Final pet state saved", "name", ui.currentPet.Name)
				}
			}
		})
	}

	// Set up a connection closed handler to ensure proper cleanup
	go func() {
		<-sessionCtx.Done()
		log.Info("SSH session closed, initiating clean shutdown")
		shutdown()
	}()

	// Ticker goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				p.Send(timeMsg(t))
			case <-ctx.Done():
				log.Debug("Time ticker stopped")
				return
			}
		}
	}()

	// Auto-save goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if ui.currentPet != nil {
					ui.syncPetState()
					err := petRepo.Save(context.Background(), ui.currentPet, ui.publicKey)
					if err != nil {
						log.Error("Error saving pet", "error", err)
					} else {
						log.Debug("Pet state auto-saved", "name", ui.currentPet.Name)
					}
				}
			case <-ctx.Done():
				log.Debug("Auto-save routine stopped")
				return
			}
		}
	}()

	// Program starter goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer shutdown() // Ensure shutdown runs even if program exits unexpectedly

		err := p.Start()
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "context canceled") ||
				strings.Contains(errStr, "program was killed") {
				log.Debug("Program exited normally", "reason", errStr)
			} else {
				log.Error("Program error", "error", err)
			}
		}
	}()

	return p
}

func getContextKeys(ctx context.Context) []string {
	keys := []string{}

	if ctx.Value(db.ContextKeyDB) != nil {
		keys = append(keys, "db.ContextKeyDB")
	}

	if ctx.Value(string(PublicKeyKey)) != nil {
		keys = append(keys, "PublicKeyKey")
	}

	for _, key := range []string{"user", "remote_addr", "session_id", "client_version", "server_version"} {
		if ctx.Value(key) != nil {
			keys = append(keys, key)
		}
	}

	return keys
}

func simulateTimePassed(p *pet.Pet, duration time.Duration) {
	hours := duration.Hours()

	log.Info("Simulating time passage", "hours", hours)

	// Cap at 7 days for simulation
	simulatedHours := hours
	if simulatedHours > 168 {
		log.Info("Capping simulation at 7 days", "actual_hours", hours)
		simulatedHours = 168
	}

	// Calculate days for easier reasoning
	days := simulatedHours / 24

	hungerPerDay := 35.0
	hungerIncrease := int(days * hungerPerDay)
	p.Hunger += hungerIncrease
	if p.Hunger > 100 {
		p.Hunger = 100
	}

	happinessLossPerDay := 35.0
	happinessDecrease := int(days * happinessLossPerDay)
	p.Happiness -= happinessDecrease
	if p.Happiness < 0 {
		p.Happiness = 0
	}

	healthDecrease := 0
	if days < 2 {
		healthDecrease = int(days * 5)
	} else if days < 4 {
		healthDecrease = 10 + int((days-2)*10)
	} else if days < 5 {
		healthDecrease = 30 + int((days-4)*20)
	} else if days < 6 {
		healthDecrease = 50 + int((days-5)*30)
	} else {
		healthDecrease = 80 + int((days-6)*40)
	}

	if p.Hunger > 80 {
		// No penalty for first 3 days
		hungerDays := math.Max(0, days-3)
		healthDecrease += int(hungerDays * 5)
	}

	// Poop guaranteed after 2 days
	if !p.HasPooped {
		if days > 2 || rand.Float64() < (days/2) {
			p.HasPooped = true
		}
	}

	if p.HasPooped {
		poopDays := math.Max(0, days-2)
		healthDecrease += int(poopDays * 3)
	}

	if !p.IsSick {
		sickChance := days / 3
		if days > 3 || rand.Float64() < sickChance {
			p.IsSick = true
		}
	}

	if p.IsSick {
		sickDays := math.Max(0, days-3)
		healthDecrease += int(sickDays * 7)
	}

	// Kill after 7 days absent
	if hours > 168 {
		healthDecrease = 100
	}

	p.Health -= healthDecrease
	if p.Health < 0 {
		p.Health = 0
	}

	log.Info("Time simulation results",
		"hunger", p.Hunger,
		"happiness", p.Happiness,
		"health", p.Health,
		"health_decrease", healthDecrease,
		"days_absent", days,
		"is_sick", p.IsSick,
		"has_pooped", p.HasPooped)
}
