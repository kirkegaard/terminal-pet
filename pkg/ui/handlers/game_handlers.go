package handlers

import (
	"context"
	"math/rand"
	"time"

	"github.com/charmbracelet/log"
	"github.com/kirkegaard/terminal-pet/pkg/db/repo"
	"github.com/kirkegaard/terminal-pet/pkg/pet"
)

func StartGame() (inGame bool, guessesLeft int, score int, number int, showResult bool, lastNumber int, lastGuessWasCorrect bool, animState string) {
	inGame = true
	guessesLeft = 5
	score = 0
	number = rand.Intn(9) + 1
	showResult = false
	lastNumber = 0
	lastGuessWasCorrect = false
	animState = "playing"

	return
}

func HandleGameGuess(
	guessHigher bool,
	inGame bool,
	gameNumber int,
	gameGuessesLeft int,
	gameScore int,
	showResult bool,
	lastGuessWasCorrect bool,
	lastNumber int,
	p *pet.Pet,
) (bool, int, int, int, bool, bool, int, string, *pet.Pet) {
	newInGame := inGame
	newGameNumber := gameNumber
	newGameGuessesLeft := gameGuessesLeft - 1
	newGameScore := gameScore
	newShowResult := true
	newLastGuessWasCorrect := false
	newLastNumber := gameNumber
	newAnimState := "playing"

	rand.Seed(time.Now().UnixNano())

	var nextNumber int
	for {
		nextNumber = rand.Intn(9) + 1
		if nextNumber != gameNumber {
			break
		}
	}

	guessCorrect := false

	if (gameNumber == 1 && guessHigher) || (gameNumber == 9 && !guessHigher) {
		guessCorrect = true
	} else if guessHigher {
		if nextNumber > gameNumber {
			guessCorrect = true
		}
	} else {
		if nextNumber < gameNumber {
			guessCorrect = true
		}
	}

	// Don't let pet get too thin
	p.Weight -= 1
	if p.Weight < 10 {
		p.Weight = 10
	}

	if guessCorrect {
		newGameScore++
		newLastGuessWasCorrect = true
		newAnimState = "happy"

		p.Happiness += 2
		if p.Happiness > 100 {
			p.Happiness = 100
		}
	} else {
		newLastGuessWasCorrect = false
		newAnimState = "sad"

		p.Happiness -= 1
		if p.Happiness < 0 {
			p.Happiness = 0
		}
	}

	if newGameGuessesLeft <= 0 || newGameScore >= 5 {
		newInGame = false

		finalHappinessBoost := newGameScore * 5
		p.Happiness += finalHappinessBoost
		if p.Happiness > 100 {
			p.Happiness = 100
		}

		if newGameScore >= 5 {
			newAnimState = "happy"
		} else if newGameScore >= 3 {
			newAnimState = "happy"
		} else {
			newAnimState = "idle"
		}

		return newInGame, newGameNumber, newGameGuessesLeft, newGameScore, newShowResult, newLastGuessWasCorrect, newLastNumber, newAnimState, p
	}

	newGameNumber = nextNumber

	return newInGame, newGameNumber, newGameGuessesLeft, newGameScore, newShowResult, newLastGuessWasCorrect, newLastNumber, newAnimState, p
}

func HandleGameOver(msg string, gameOverCursor int) (int, bool) {
	switch msg {
	case "up", "down", "k", "j":
		return 1 - gameOverCursor, false
	case "enter", " ":
		if gameOverCursor == 0 {
			return gameOverCursor, true
		} else {
			return gameOverCursor, false
		}
	default:
		return gameOverCursor, false
	}
}

func RestartGame(name string, parent *pet.Parent) *pet.Pet {
	log.Debug("Restarting game")

	// Create a new pet with default values
	newPet := pet.NewPet(name, time.Now(), parent)

	// Preserve the parent ID which is needed for database operations
	if parent != nil && parent.ID > 0 {
		log.Debug("Using existing parent ID", "parentID", parent.ID)
	}

	// Get the pet repository and persist the new pet
	petRepo := repo.NewPetRepository(nil)
	err := petRepo.Create(context.Background(), newPet)
	if err != nil {
		log.Error("Failed to create pet in database", "error", err)
		// If there's an error, just return the pet without persistence
	} else {
		log.Debug("Created new pet in database", "id", newPet.ID, "name", newPet.Name)
	}

	return newPet
}
