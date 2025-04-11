package handlers

import (
	"github.com/kirkegaard/terminal-pet/pkg/pet"
	"github.com/kirkegaard/terminal-pet/pkg/pet/ascii"
)

// HandleDebugMenu processes input in the debug menu
func HandleDebugMenu(key string, cursor int, itemCount int) (stayInMenu bool, newCursor int) {
	newCursor = cursor

	switch key {
	case "esc", "b", "q":
		// Exit debug menu
		return false, cursor
	case "up", "k":
		// Move cursor up
		if cursor > 0 {
			newCursor = cursor - 1
		}
	case "down", "j":
		// Move cursor down
		if cursor < itemCount-1 {
			newCursor = cursor + 1
		}
	}

	return true, newCursor
}

// ExecuteDebugAction performs the selected debug action
func ExecuteDebugAction(
	debugCursor int,
	p *pet.Pet,
	debugMode bool,
	inDebugMenu bool,
	inGameOver bool,
	gameOverCursor int,
) (*pet.Pet, bool, bool, bool, int, ascii.Animation) {
	switch debugCursor {
	case 0: // Toggle Sick
		p.IsSick = !p.IsSick
	case 1: // Toggle Poop
		p.HasPooped = !p.HasPooped
	case 2: // Toggle Dead
		if p.Health <= 0 {
			p.Health = 50 // Revive
		} else {
			p.Health = 0 // Kill
			inGameOver = true
			gameOverCursor = 0
		}
	case 3: // Set Full Health (100)
		p.Health = 100
	case 4: // Set Low Health (30)
		p.Health = 30
	case 5: // Set Not Hungry (0)
		p.Hunger = 0
	case 6: // Set Very Hungry (90)
		p.Hunger = 90
	case 7: // Set Happy (100)
		p.Happiness = 100
	case 8: // Set Sad (10)
		p.Happiness = 10
	case 9: // Critical State
		p.Health = 1
		p.Hunger = 100
		p.Happiness = 0
	case 10: // Reset All Stats
		p.Health = 100
		p.Hunger = 0
		p.Happiness = 100
		p.Weight = 50
		p.IsSick = false
		p.HasPooped = false
	case 11: // Exit Debug Mode
		debugMode = false
		inDebugMenu = false
	case 12: // Set Obese (Weight=110)
		p.Weight = 110
	case 13: // Set Normal Weight (Weight=50)
		p.Weight = 50
	case 14: // Set Athletic (Weight=20)
		p.Weight = 20
	}

	// Get the updated animation based on the new pet state
	newAnim := ascii.GetAnimationForState(p.GetState())

	return p, debugMode, inDebugMenu, inGameOver, gameOverCursor, newAnim
}
