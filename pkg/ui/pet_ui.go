package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/kirkegaard/terminal-pet/pkg/pet"
	"github.com/kirkegaard/terminal-pet/pkg/pet/ascii"
	"github.com/kirkegaard/terminal-pet/pkg/ui/handlers"
	"github.com/kirkegaard/terminal-pet/pkg/ui/keymap"
	"github.com/kirkegaard/terminal-pet/pkg/ui/views"
)

// QuitMsg is a custom message used to signal a quit request from the menu
type QuitMsg struct{}

type FrameMsg time.Time

const (
	AnimationTickRate = time.Second / 2
	ResultDisplayTime = 1000
)

type GameResultTimeoutMsg struct{}

var choices = []string{"Feed", "Clean", "Play", "Medicine", "Rename", "Toggle Lights", "Quit"}

var infoStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#888888")).
	Bold(false).
	Italic(true)

type PetUI struct {
	pet                *pet.Pet
	currentAnim        ascii.Animation
	currentFrame       int
	lastUpdateTime     time.Time
	lastStatUpdateTime time.Time // Track when stats were last updated
	cursor             int
	selectedAction     int
	selectedTime       time.Time
	keys               keymap.KeyMap
	help               help.Model
	width              int
	height             int
	showHelp           bool
	showStats          bool

	// Animation state
	animState      string
	frameCounter   int
	animCompleted  bool
	petPosition    int
	targetPosition int
	moveDirection  int

	// Game state
	inGame          bool
	gameNumber      int
	gameGuessesLeft int
	gameScore       int

	// Game result display
	lastGuessWasCorrect bool
	lastNumber          int
	showResult          bool

	// Game over state
	inGameOver       bool
	gameOverCursor   int
	restartRequested bool
	justRestarted    bool

	// Debug mode
	debugMode   bool
	inDebugMenu bool
	debugCursor int

	// Rename mode
	inRenameMode bool
	newName      string

	// Food selection mode
	inFoodSelectMode bool
	foodCursor       int
	foodOptions      []string

	// Inline food submenu
	showFoodSubmenu   bool
	foodSubmenuCursor int
}

// GetPet returns the pet reference
func (m *PetUI) GetPet() *pet.Pet {
	return m.pet
}

// ResetRestartFlag resets the restart flag after saving
func (m *PetUI) ResetRestartFlag() {
	m.justRestarted = false
}

// SetGameOver sets the game over state to the specified value
func (m *PetUI) SetGameOver(isGameOver bool) {
	m.inGameOver = isGameOver
	if isGameOver {
		m.gameOverCursor = 0 // Default to restart option
	}
}

// NewPetUI creates a new pet UI
func NewPetUI(p *pet.Pet, width, height int) *PetUI {
	anim := ascii.GetAnimationForState(p.GetState())

	// Check if pet is already dead when loading and set initial game over state
	inGameOver := p.IsDead()
	initialCursor := 0
	if inGameOver {
		anim = ascii.GetAnimationForState(ascii.StateDead)
	}

	now := time.Now()

	// Initialize help model with ShowAll set to true to display all keys
	helpModel := help.New()
	helpModel.ShowAll = true

	return &PetUI{
		pet:                p,
		currentAnim:        anim,
		currentFrame:       0,
		keys:               keymap.Keys,
		help:               helpModel,
		width:              width,
		height:             height,
		showHelp:           true,
		showStats:          true,
		lastUpdateTime:     now,
		lastStatUpdateTime: now,
		cursor:             0,
		selectedAction:     -1, // -1 means no selection
		selectedTime:       time.Time{},

		// Initialize animation state
		animState:      "idle", // Start in idle state
		frameCounter:   0,
		animCompleted:  false,
		petPosition:    0,
		targetPosition: 0,
		moveDirection:  0,

		// Game state
		inGame:              false,
		gameNumber:          0,
		gameGuessesLeft:     0,
		gameScore:           0,
		lastGuessWasCorrect: false,
		lastNumber:          0,
		showResult:          false,
		debugMode:           false,
		inDebugMenu:         false,
		debugCursor:         0,

		// Game over state - initialize from pet status
		inGameOver:       inGameOver,
		gameOverCursor:   initialCursor,
		restartRequested: false,

		// Rename mode
		inRenameMode: false,
		newName:      "",

		// Food selection mode
		inFoodSelectMode: false,
		foodCursor:       0,
		foodOptions:      []string{"Burger", "Cake"},

		// Inline food submenu
		showFoodSubmenu:   false,
		foodSubmenuCursor: 0,
	}
}

// Init initializes the model
func (m *PetUI) Init() tea.Cmd {
	return m.startGlobalTicker()
}

// startGlobalTicker creates the single global ticker that powers everything
func (m *PetUI) startGlobalTicker() tea.Cmd {
	return tea.Tick(AnimationTickRate, func(t time.Time) tea.Msg {
		return FrameMsg(t)
	})
}

// resetToIdle sets the animation state back to idle based on current pet state
func (m *PetUI) resetToIdle() {
	m.animState = "idle"
	m.currentAnim = ascii.GetAnimationForState(m.pet.GetState())
	m.currentFrame = 0
	m.frameCounter = 0
}

// updateAnimation handles all animation state transitions and frame updates
func (m *PetUI) updateAnimation(now time.Time) {
	// Check if the pet just died
	if m.pet.IsDead() && !m.inGameOver {
		m.inGameOver = true
		m.gameOverCursor = 0
		m.currentAnim = ascii.GetAnimationForState(ascii.StateDead)
		m.animState = "dead"
		m.currentFrame = 0
		return
	}

	// Game-specific animations take priority
	if m.inGame {
		// Don't change animation during game unless explicitly requested
		return
	}

	// Process animation state transitions based on current state
	switch m.animState {
	case "idle":
		newAnim := ascii.GetAnimationForState(m.pet.GetState())

		if newAnim.Name != m.currentAnim.Name {
			m.currentAnim = newAnim
			m.currentFrame = 0
			m.frameCounter = 0
		}

	case "eating", "cakeEating":
		if m.frameCounter >= len(m.currentAnim.Frames) {
			m.resetToIdle()
		}

	case "happy", "sad":
		if m.frameCounter >= len(m.currentAnim.Frames)*2 {
			m.resetToIdle()
		}

	case "playing":
		// Playing animation is handled by game logic

	case "dead":
		m.currentAnim = ascii.GetAnimationForState(ascii.StateDead)
	}

	if len(m.currentAnim.Frames) > 0 {
		m.currentFrame = (m.currentFrame + 1) % len(m.currentAnim.Frames)
		m.frameCounter++
	}
}

func (m *PetUI) handlePetMovement() {
	// Only move the pet in these states
	canMove := m.currentAnim.Name == "Idle" || m.currentAnim.Name == "Happy"

	if !canMove || m.pet.IsDead() {
		if m.petPosition > 0 {
			m.petPosition--
		} else if m.petPosition < 0 {
			m.petPosition++
		}
		return
	}

	if m.moveDirection == 0 {
		if rand.Intn(2) == 0 {
			m.moveDirection = 1
		} else {
			m.moveDirection = -1
		}
	}

	// Movement is determined by a simple sine wave pattern
	// We'll change our position every 3 frames
	if m.frameCounter%3 == 0 {
		if rand.Intn(10) == 0 {
			m.moveDirection *= -1
		}

		m.petPosition += m.moveDirection

		if m.petPosition > 3 {
			m.petPosition = 3
			m.moveDirection = -1
		} else if m.petPosition < -3 {
			m.petPosition = -3
			m.moveDirection = 1
		}
	}
}

func (m *PetUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Check for restart request
	if m.restartRequested {
		log.Debug("Restarting game")
		return m.restartGame()
	}

	switch msg := msg.(type) {
	case FrameMsg:
		// Global ticker handles all animations and state transitions

		// Handle result display timeout
		if m.inGame && m.showResult {
			// Get time since last frame
			now := time.Now()
			elapsed := now.Sub(m.lastUpdateTime).Milliseconds()

			// If we've displayed the result long enough, advance the game
			if elapsed >= ResultDisplayTime {
				m.showResult = false
				m.animState = "playing"
				m.currentAnim = ascii.Playing
				m.currentFrame = 0
				m.frameCounter = 0
				m.lastUpdateTime = now
			}
		} else {
			now := time.Now()
			m.lastUpdateTime = now

			m.updateAnimation(now)

			m.handlePetMovement()

			if now.Sub(m.lastStatUpdateTime) >= time.Second && !m.debugMode {
				m.updatePetState()
				m.lastStatUpdateTime = now
			}

			if m.selectedAction >= 0 && time.Since(m.selectedTime) > 1*time.Second {
				m.selectedAction = -1
			}
		}

		return m, m.startGlobalTicker()

	case tea.KeyMsg:
		// Handle game over screen if active
		if m.inGameOver {
			newCursor, shouldRestart := handlers.HandleGameOver(msg.String(), m.gameOverCursor)
			m.gameOverCursor = newCursor

			if shouldRestart {
				m.restartRequested = true
				return m, nil
			} else if newCursor == 1 && msg.String() == "enter" {
				return m, tea.Quit
			}
			return m, nil
		}

		// Debug mode toggle with Ctrl+D
		if msg.String() == "ctrl+d" {
			m.debugMode = !m.debugMode
			m.inDebugMenu = m.debugMode
			return m, nil
		}

		// Handle debug menu if active
		if m.inDebugMenu {
			stayInMenu, newCursor := handlers.HandleDebugMenu(msg.String(), m.debugCursor, views.GetDebugMenuItemCount())
			m.debugCursor = newCursor

			if !stayInMenu {
				m.inDebugMenu = false
				return m, nil
			}

			if msg.String() == "enter" || msg.String() == " " {
				m.pet, m.debugMode, m.inDebugMenu, m.inGameOver, m.gameOverCursor, m.currentAnim = handlers.ExecuteDebugAction(
					m.debugCursor,
					m.pet,
					m.debugMode,
					m.inDebugMenu,
					m.inGameOver,
					m.gameOverCursor,
				)

				// Reset animation state
				m.resetToIdle()
			}

			return m, nil
		}

		// Handle food selection mode if active
		if m.inFoodSelectMode {
			stayInFoodMode, newCursor, selected := handlers.HandleFoodSelection(msg.String(), m.foodCursor, len(m.foodOptions))
			m.foodCursor = newCursor

			if !stayInFoodMode {
				m.inFoodSelectMode = false

				if selected {
					animState, updatedPet := handlers.FeedPet(m.foodCursor, m.pet)
					m.pet = updatedPet
					m.animState = animState

					// Set correct animation based on state
					if animState == "eating" {
						m.currentAnim = ascii.Eating
					} else if animState == "cakeEating" {
						m.currentAnim = ascii.CakeEating
					}
					m.currentFrame = 0
					m.frameCounter = 0
				}
			}

			return m, nil
		}

		// If we're in a game, handle game controls
		if m.inGame {
			switch msg.String() {
			case "h", "left":
				// Player guesses "lower"
				wasInGame := m.inGame
				m.inGame, m.gameNumber, m.gameGuessesLeft, m.gameScore, m.showResult, m.lastGuessWasCorrect, m.lastNumber, m.animState, m.pet = handlers.HandleGameGuess(
					false,
					m.inGame,
					m.gameNumber,
					m.gameGuessesLeft,
					m.gameScore,
					m.showResult,
					m.lastGuessWasCorrect,
					m.lastNumber,
					m.pet,
				)

				// Update animation based on state
				if m.animState == "happy" {
					m.currentAnim = ascii.Happy
				} else if m.animState == "sad" {
					m.currentAnim = ascii.Sad
				} else if m.animState == "playing" {
					m.currentAnim = ascii.Playing
				} else if m.animState == "idle" {
					m.currentAnim = ascii.GetAnimationForState(m.pet.GetState())
				}

				// Reset frames
				m.currentFrame = 0
				m.frameCounter = 0
				m.lastUpdateTime = time.Now()

				// If game just ended, setup idle state
				if !m.inGame && wasInGame {
					m.resetToIdle()
					m.showResult = false
				}

				return m, nil

			case "l", "right":
				// Player guesses "higher"
				wasInGame := m.inGame
				m.inGame, m.gameNumber, m.gameGuessesLeft, m.gameScore, m.showResult, m.lastGuessWasCorrect, m.lastNumber, m.animState, m.pet = handlers.HandleGameGuess(
					true,
					m.inGame,
					m.gameNumber,
					m.gameGuessesLeft,
					m.gameScore,
					m.showResult,
					m.lastGuessWasCorrect,
					m.lastNumber,
					m.pet,
				)

				// Update animation based on state
				if m.animState == "happy" {
					m.currentAnim = ascii.Happy
				} else if m.animState == "sad" {
					m.currentAnim = ascii.Sad
				} else if m.animState == "playing" {
					m.currentAnim = ascii.Playing
				} else if m.animState == "idle" {
					m.currentAnim = ascii.GetAnimationForState(m.pet.GetState())
				}

				// Reset frames
				m.currentFrame = 0
				m.frameCounter = 0
				m.lastUpdateTime = time.Now()

				// If game just ended, setup idle state
				if !m.inGame && wasInGame {
					m.resetToIdle()
					m.showResult = false
				}

				return m, nil

			case "esc":
				// Exit game
				m.inGame = false
				m.resetToIdle()
				m.showResult = false

				return m, nil

			default:
				// Even for unhandled keys, maintain the animation timer
				return m, m.startGlobalTicker()
			}
		}

		// Normal UI controls when not in game
		switch {
		case key.Matches(msg, m.keys.Quit):
			// Use the custom quit message instead of tea.Quit
			return m, func() tea.Msg { return QuitMsg{} }

		// Add this case for toggling help visibility
		case key.Matches(msg, m.keys.Help):
			// Toggle help visibility
			m.showHelp = !m.showHelp
			return m, nil

		// Back button - ESC
		case msg.String() == "esc":
			// ESC acts as a back/cancel action in various screens
			if m.inRenameMode {
				m.inRenameMode = false
				m.newName = ""
				return m, nil
			} else if m.inFoodSelectMode {
				m.inFoodSelectMode = false
				return m, nil
			} else if m.inDebugMenu {
				m.inDebugMenu = false
				return m, nil
			} else if m.showFoodSubmenu {
				m.showFoodSubmenu = false
				return m, nil
			}
			// On the main screen, ESC does nothing
			return m, nil

		// Debug shortcut to force game over
		case msg.String() == "ctrl+k":
			m.pet.Health = 0
			m.inGameOver = true
			m.gameOverCursor = 0
			return m, nil

		// Handle rename mode input
		case m.inRenameMode:
			switch msg.String() {
			// Q is already handled by the Quit key binding above
			// ESC is already handled above
			case "enter", "return":
				// Confirm rename
				if len(m.newName) > 0 {
					m.pet.Name = m.newName
					m.inRenameMode = false
					m.newName = ""
				}
			case "backspace":
				// Delete last character
				if len(m.newName) > 0 {
					m.newName = m.newName[:len(m.newName)-1]
				}
			default:
				// Only accept printable characters and limit to reasonable length
				if len(msg.String()) == 1 && len(m.newName) < 20 {
					r := []rune(msg.String())[0]
					if unicode.IsPrint(r) {
						m.newName += msg.String()
					}
				}
			}
			// Always return the tick command to keep the cursor blinking
			return m, m.startGlobalTicker()

		case key.Matches(msg, m.keys.Action):
			if m.showFoodSubmenu {
				// Handle food submenu selection
				keepSubmenu, newCursor := handlers.HandleFoodSubmenu(msg.String(), m.foodSubmenuCursor, len(m.foodOptions))

				if !keepSubmenu {
					m.showFoodSubmenu = false
					return m, nil
				}

				m.foodSubmenuCursor = newCursor

				if msg.String() == "enter" || msg.String() == " " {
					m.showFoodSubmenu = false
					animState, updatedPet := handlers.FeedPet(m.foodSubmenuCursor, m.pet)
					m.pet = updatedPet
					m.animState = animState

					// Set correct animation based on state
					if animState == "eating" {
						m.currentAnim = ascii.Eating
					} else if animState == "cakeEating" {
						m.currentAnim = ascii.CakeEating
					}
					m.currentFrame = 0
					m.frameCounter = 0
				}

				return m, nil
			} else {
				// Set the selected action for highlighting
				m.selectedAction = m.cursor
				m.selectedTime = time.Now()

				// If pet is sleeping (lights off), only allow toggling lights or quitting
				if !m.pet.LightsOn && m.cursor != 5 && m.cursor != 6 { // 5 is Toggle Lights, 6 is Quit
					// Pet is sleeping, can't perform other actions
					return m, nil
				}

				// Regular action handling if not in food submenu
				if m.cursor == 0 { // Feed
					// Show food submenu instead of food select screen
					m.showFoodSubmenu = true
					m.foodSubmenuCursor = 0 // Reset to first option
				} else if m.cursor == 1 { // Clean
					m.pet.Clean()
					// Could add a cleaning animation here in the future
				} else if m.cursor == 2 { // Play
					// Start the game with proper timer initialization
					return m.startGame()
				} else if m.cursor == 3 { // Medicine
					m.pet.GiveMedicine()
					// Could add a medicine animation here in the future
				} else if m.cursor == 4 { // Rename
					// Implement rename functionality
					m.inRenameMode = true
				} else if m.cursor == 5 { // Toggle Lights
					// Implement toggle lights functionality
					m.pet.ToggleLights()
				} else if m.cursor == 6 { // Quit
					// Instead of directly quitting, emit a custom quit message that will allow
					// parent components to save state before quitting
					return m, func() tea.Msg { return QuitMsg{} }
				}
			}
		case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Left):
			if m.showFoodSubmenu {
				// Move left in food submenu
				if msg.String() == "left" {
					if m.foodSubmenuCursor > 0 {
						m.foodSubmenuCursor--
					}
				}
			} else if m.cursor > 0 {
				// If pet is sleeping, only allow moving to Toggle Lights or Quit
				if !m.pet.LightsOn {
					// Only allow movement to Toggle Lights (5) or Quit (6)
					if m.cursor > 6 {
						m.cursor--
					} else if m.cursor == 6 {
						m.cursor = 5
					} else if m.cursor < 5 {
						m.cursor = 5
					}
				} else {
					m.cursor--
				}
			}
		case key.Matches(msg, m.keys.Down), key.Matches(msg, m.keys.Right):
			if m.showFoodSubmenu {
				// Handle right key in food submenu
				if msg.String() == "right" {
					if m.foodSubmenuCursor < len(m.foodOptions)-1 {
						m.foodSubmenuCursor++
					}
				}
				// Down key doesn't do anything in food submenu
			} else if m.cursor < len(choices)-1 {
				// If pet is sleeping, only allow moving to Toggle Lights or Quit
				if !m.pet.LightsOn {
					// Only allow movement to Toggle Lights (5) or Quit (6)
					if m.cursor < 5 {
						m.cursor = 5
					} else if m.cursor == 5 {
						m.cursor = 6
					} else if m.cursor > 6 {
						m.cursor = 6
					}
				} else {
					m.cursor++
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
	}

	return m, cmd
}

// updatePetState updates the pet state over time
func (m *PetUI) updatePetState() {
	// If pet is dead, don't update stats
	if m.pet.IsDead() {
		return
	}

	// Special sleep benefits when lights are off
	if !m.pet.LightsOn {
		// When sleeping (lights off), pet recovers health and happiness
		if m.pet.Health < 100 && rand.Intn(3) == 0 {
			m.pet.Health += 1
		}

		if m.pet.Happiness < 100 && rand.Intn(5) == 0 {
			m.pet.Happiness += 1
		}

		// Hunger increases more slowly when sleeping
		if m.pet.Hunger < 100 && rand.Intn(10) == 0 {
			m.pet.Hunger += 1
		}

		// Cap values
		if m.pet.Health > 100 {
			m.pet.Health = 100
		}
		if m.pet.Happiness > 100 {
			m.pet.Happiness = 100
		}
		if m.pet.Hunger > 100 {
			m.pet.Hunger = 100
		}

		// Early return - when sleeping no other effects apply
		return
	}

	// Normal state updates (when lights are on)
	// Increase hunger over time, but only if pet is not full
	// Make hunger increase more consistently for testing
	if m.pet.Hunger < 100 && rand.Intn(5) == 0 {
		m.pet.Hunger += 1
	}

	// Cap hunger at 100
	if m.pet.Hunger > 100 {
		m.pet.Hunger = 100
	}

	// Pet loses health when very hungry - always check this regardless of other stats
	if m.pet.Hunger > 90 {
		m.pet.Health -= 1
	}

	// Ensure hunger never goes below 0
	if m.pet.Hunger < 0 {
		m.pet.Hunger = 0
	}

	// Decrease happiness over time (increased rate for testing)
	if rand.Intn(2) == 0 {
		m.pet.Happiness -= 1
	}

	// Cap happiness between 0 and 100
	if m.pet.Happiness > 100 {
		m.pet.Happiness = 100
	}
	if m.pet.Happiness < 0 {
		m.pet.Happiness = 0
	}

	// Health slowly recovers if not too hungry
	if m.pet.Hunger < 80 && m.pet.Health < 100 && !m.pet.IsSick {
		m.pet.Health += 1
	}

	// Health gradually decreases over time - slowed down to once every 3 minutes on average
	if rand.Intn(180) == 0 {
		m.pet.Health -= 1
	}

	// Weight affects health - heavier pets lose health faster
	if m.pet.Weight > 100 {
		if rand.Intn(60) == 0 {
			m.pet.Health -= 1
		}
	} else if m.pet.Weight > 75 {
		if rand.Intn(120) == 0 {
			m.pet.Health -= 1
		}
	}

	// Cap health between 0 and 100
	if m.pet.Health > 100 {
		m.pet.Health = 100
	}
	if m.pet.Health < 0 {
		m.pet.Health = 0
		// If health reaches 0, set game over state
		m.inGameOver = true
		m.gameOverCursor = 0
	}

	// Random chance for sickness (if not already sick)
	if !m.pet.IsSick && !m.pet.IsDead() {
		if rand.Intn(3600) == 0 {
			m.pet.IsSick = true
		}
	}

	// Natural weight loss over time when hungry
	if m.pet.Hunger > 50 && rand.Intn(300) == 0 {
		m.pet.Weight -= 1
		if m.pet.Weight < 10 {
			m.pet.Weight = 10
		}
	}

	// Sickness decreases health
	if m.pet.IsSick {
		// Health decreases faster when sick
		if rand.Intn(20) == 0 {
			m.pet.Health -= 1
		}
	}

	// Random chance to poop (if not already pooped)
	if !m.pet.HasPooped && !m.pet.IsDead() {
		// Base chance to poop
		poopChance := 1800 // Default ~once per 30 minutes

		// Babies poop much more frequently
		if m.pet.LifeStage() == pet.StageBaby {
			poopChance = 300 // Babies poop ~6x more often (every 5 minutes)
		} else if m.pet.LifeStage() == pet.StageChild {
			poopChance = 900 // Children poop ~2x more often (every 15 minutes)
		}

		// Recently fed pets poop more
		if m.pet.Hunger < 30 {
			poopChance = poopChance / 2 // Twice as likely to poop when recently fed
		}

		if rand.Intn(poopChance) == 0 {
			m.pet.HasPooped = true
		}
	}

	// Uncleaned poop gradually decreases health and happiness
	if m.pet.HasPooped {
		if rand.Intn(30) == 0 {
			m.pet.Health -= 1
			m.pet.Happiness -= 1
		}
	}
}

// View renders the UI
func (m *PetUI) View() string {
	var output string

	// If we're in debug menu, render it instead of the normal view
	if m.inDebugMenu {
		output = views.RenderDebugMenu(
			"",
			m.width,
			m.pet,
			m.debugCursor,
		)
	} else if m.inGameOver {
		// If we're in game over, render the game over screen
		output = views.RenderGameOver(
			"",
			m.width,
			m.pet,
			m.gameOverCursor,
		)
	} else if m.inGame {
		// If we're in game, render the game UI
		output = views.RenderGameView(
			"",
			m.width,
			m.currentFrame,
			m.animState,
			m.petPosition,
			m.showResult,
			m.lastGuessWasCorrect,
			m.gameNumber,
			m.lastNumber,
			m.gameGuessesLeft,
			m.gameScore,
			m.inGame,
		)
	} else if m.inRenameMode {
		// If we're in rename mode, render the rename UI
		output = views.RenderRenameView(
			"",
			m.width,
			m.newName,
		)
	} else if m.inFoodSelectMode {
		// If we're in food selection mode, render the food selection UI
		output = views.RenderFoodSelectionView(
			"",
			m.width,
			m.pet,
			m.foodCursor,
			m.foodOptions,
		)
	} else {
		// Render the main view with fixed parameters to match the function signature
		output = views.RenderMainView(
			m.pet,
			m.currentAnim,
			m.currentFrame,
			m.petPosition,
			m.width,
			m.showStats,
			m.showHelp,
			m.help,
			m.keys,
			m.cursor,
			m.selectedAction,
			m.debugMode,
			m.showFoodSubmenu,
			m.foodSubmenuCursor,
			m.foodOptions,
		)
	}

	// Add debug information at the bottom if in debug mode
	if m.debugMode {
		// Fix any trailing underscore in animState
		displayState := m.animState
		displayState = strings.TrimSuffix(displayState, "_")

		// Get the expected animation based on current pet state
		expectedAnim := ascii.GetAnimationForState(m.pet.GetState())

		// Simplified animation display using the Name field
		expectedAnimName := expectedAnim.Name
		currentAnimName := m.currentAnim.Name

		// Show mismatch warning if animation state is "idle" but we're showing a different animation
		mismatchWarning := ""
		if displayState == "idle" && expectedAnimName != currentAnimName {
			mismatchWarning = " (MISMATCH!)"
		}

		debugInfo := fmt.Sprintf(
			"DEBUG: Frame: %d, State: %s, Animation: %s%s, Expected: %s, FPS: %d, Position: %d, In Game: %t",
			m.currentFrame,
			displayState,
			currentAnimName,
			mismatchWarning,
			expectedAnimName,
			m.currentAnim.FPS,
			m.petPosition,
			m.inGame,
		)

		// Add more pet state information
		petStateInfo := fmt.Sprintf(
			"PET: Health: %d, Hunger: %d, Happiness: %d, Weight: %d, Sick: %t, HasPooped: %t, Lights: %t",
			m.pet.Health,
			m.pet.Hunger,
			m.pet.Happiness,
			m.pet.Weight,
			m.pet.IsSick,
			m.pet.HasPooped,
			m.pet.LightsOn,
		)

		// Add separator
		output += "\n" + strings.Repeat("─", m.width) + "\n"

		// Add debug info
		output += infoStyle.Render(debugInfo)
		output += "\n" + infoStyle.Render(petStateInfo)

		// Add controls reminder
		output += "\n" + infoStyle.Render("Press Ctrl+D to toggle debug menu")
	}

	return output
}

// Initializes a new game
func (m *PetUI) startGame() (tea.Model, tea.Cmd) {
	m.inGame, m.gameGuessesLeft, m.gameScore, m.gameNumber, m.showResult, m.lastNumber, m.lastGuessWasCorrect, m.animState = handlers.StartGame()

	m.currentAnim = ascii.Playing

	now := time.Now()
	m.currentFrame = 0
	m.frameCounter = 0
	m.animCompleted = false
	m.lastUpdateTime = now

	return m, nil
}

// Creates a new pet and resets the game state
func (m *PetUI) restartGame() (tea.Model, tea.Cmd) {
	m.pet = handlers.RestartGame(m.pet.Name, m.pet.Parent)

	// Reset UI state
	now := time.Now()
	m.currentAnim = ascii.GetAnimationForState(m.pet.GetState())
	m.currentFrame = 0
	m.lastUpdateTime = now
	m.lastStatUpdateTime = now
	m.cursor = 0
	m.selectedAction = -1
	m.inGameOver = false
	m.restartRequested = false
	m.justRestarted = true

	// Reset animation state
	m.resetToIdle()

	// Reset game states
	m.inGame = false
	m.gameNumber = 0
	m.gameGuessesLeft = 0
	m.gameScore = 0
	m.showResult = false
	m.lastGuessWasCorrect = false
	m.lastNumber = 0

	// Reset debug state
	m.debugMode = false
	m.inDebugMenu = false
	m.debugCursor = 0

	// Animation state
	m.petPosition = 0
	m.targetPosition = 0
	m.moveDirection = 0

	// Maintain the global ticker
	return m, m.startGlobalTicker()
}
