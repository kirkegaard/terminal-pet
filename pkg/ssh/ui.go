package ssh

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/kirkegaard/terminal-pet/pkg/db/repo"
	"github.com/kirkegaard/terminal-pet/pkg/pet"
	petui "github.com/kirkegaard/terminal-pet/pkg/ui"
)

type timeMsg time.Time

type UI struct {
	Renderer   *lipgloss.Renderer
	time       time.Time
	width      int
	height     int
	petUI      tea.Model
	currentPet *pet.Pet
	publicKey  string
}

func NewUI(ctx context.Context, renderer *lipgloss.Renderer, width int, height int, p *pet.Pet, publicKey string) *UI {
	petUIModel := petui.NewPetUI(p, width, height)

	ui := &UI{
		Renderer:   renderer,
		width:      width,
		height:     height,
		time:       time.Now(),
		petUI:      petUIModel,
		currentPet: p,
		publicKey:  publicKey,
	}

	return ui
}

func (ui *UI) Init() tea.Cmd {
	return ui.petUI.Init()
}

func (ui *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case timeMsg:
		ui.time = time.Time(msg)

		if ui.time.Second()%30 == 0 {
			ui.syncPetState()
		}

	case tea.WindowSizeMsg:
		ui.height = msg.Height
		ui.width = msg.Width
		ui.petUI, cmd = ui.petUI.Update(msg)
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			ui.syncPetState()
			return ui, tea.Quit
		}

		oldPetUI, isPetUI := ui.petUI.(*petui.PetUI)
		wasRestarted := isPetUI && oldPetUI.JustRestarted()

		ui.petUI, cmd = ui.petUI.Update(msg)

		newPetUI, isPetUI := ui.petUI.(*petui.PetUI)
		if isPetUI && !wasRestarted && newPetUI.JustRestarted() {
			ui.currentPet = newPetUI.GetPet()
			log.Info("Pet was restarted and will be saved", "name", ui.currentPet.Name)

			petRepo := repo.NewPetRepository(nil)
			err := petRepo.Save(context.Background(), ui.currentPet, ui.publicKey)
			if err != nil {
				log.Error("Error saving pet after restart", "error", err)
			} else {
				log.Info("Pet saved after restart", "name", ui.currentPet.Name)
			}

			newPetUI.ResetRestartFlag()
		}

		if cmdName := fmt.Sprintf("%T", cmd); strings.Contains(cmdName, "Quit") {
			return ui, cmd
		}

		if msg.String() == "enter" || msg.String() == " " {
			petUI, isPetUI := ui.petUI.(*petui.PetUI)
			if isPetUI {
				ui.currentPet = petUI.GetPet()
			}
		}

	default:
		ui.petUI, cmd = ui.petUI.Update(msg)

		if cmdName := fmt.Sprintf("%T", cmd); strings.Contains(cmdName, "Quit") {
			return ui, cmd
		}
	}

	return ui, cmd
}

func (ui *UI) syncPetState() {
	// Update visit time
	ui.currentPet.LastVisit = time.Now()

	// Always get the current pet state from the UI model first
	if petUIModel, ok := ui.petUI.(*petui.PetUI); ok {
		ui.currentPet = petUIModel.GetPet()
	}

	log.Debug("Pet state synced",
		"name", ui.currentPet.Name,
		"hunger", ui.currentPet.Hunger,
		"happiness", ui.currentPet.Happiness,
		"health", ui.currentPet.Health,
		"is_sick", ui.currentPet.IsSick,
		"has_pooped", ui.currentPet.HasPooped,
		"is_dead", ui.currentPet.Health <= 0)

	petRepo := repo.NewPetRepository(nil)
	err := petRepo.Save(context.Background(), ui.currentPet, ui.publicKey)
	if err != nil {
		log.Error("Error saving pet state", "error", err)
	}
}

func (ui *UI) View() string {
	return ui.petUI.View()
}
