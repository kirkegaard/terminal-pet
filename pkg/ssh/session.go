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

		existingPet.LastVisit = time.Now()

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

	// Base rates
	hungerPerHour := 2.0 + rand.Float64()*4
	happinessLossPerHour := 1.0 + rand.Float64()*1.5

	// Progressive multipliers based on duration
	durationMultiplier := 1.0
	if hours > 24 {
		durationMultiplier = 1.2 // After 1 day, things get worse
	}
	if hours > 48 {
		durationMultiplier = 1.5 // After 2 days, even worse
	}
	if hours > 72 {
		durationMultiplier = 2.0 // After 3 days, much worse
	}
	if hours > 120 {
		durationMultiplier = 3.0 // After 5 days, critical
	}

	// Apply multipliers
	hungerIncrease := int(hours * hungerPerHour * durationMultiplier)
	happinessDecrease := int(hours * happinessLossPerHour * durationMultiplier)

	// Base health decrease (gets worse with time)
	healthDecrease := int(hours * 0.2 * durationMultiplier)

	// Apply changes to pet stats
	p.Hunger += hungerIncrease
	if p.Hunger > 100 {
		extraHunger := p.Hunger - 100
		healthDecrease += extraHunger / 3 // More severe health impact from hunger
		p.Hunger = 100
	}

	p.Happiness -= happinessDecrease
	if p.Happiness < 0 {
		p.Happiness = 0
	}

	if p.Hunger > 80 {
		healthDecrease += int(hours * 0.8 * durationMultiplier)
	}

	// Guaranteed death threshold (after about a week)
	if hours > 160 {
		healthDecrease = max(healthDecrease, p.Health) // Ensure pet dies
	}

	// Pooping probability increases with time
	poopChance := hours / 12
	if hours > 48 {
		poopChance = math.Min(1.0, poopChance*1.5) // Higher chance after 2 days
	}
	if !p.HasPooped && rand.Float64() < poopChance {
		p.HasPooped = true
	}

	// Sickness probability increases with time
	sickChance := hours / 48
	if hours > 72 {
		sickChance = math.Min(1.0, sickChance*2) // Higher chance after 3 days
	}
	if !p.IsSick && rand.Float64() < sickChance {
		p.IsSick = true
	}

	// Additional effects if sick or has pooped
	if p.IsSick {
		sickHealthLoss := int((hours + 0.5) * durationMultiplier)
		healthDecrease += sickHealthLoss
	}

	if p.HasPooped {
		poopHealthLoss := int((hours*0.5 + 0.5) * durationMultiplier)
		healthDecrease += poopHealthLoss
		p.Happiness -= poopHealthLoss
		if p.Happiness < 0 {
			p.Happiness = 0
		}
	}

	// Apply total health decrease
	p.Health -= healthDecrease
	if p.Health < 0 {
		p.Health = 0
	}

	log.Info("Time simulation results",
		"hunger", p.Hunger,
		"happiness", p.Happiness,
		"health", p.Health,
		"is_sick", p.IsSick,
		"has_pooped", p.HasPooped,
		"duration_multiplier", durationMultiplier)
}
