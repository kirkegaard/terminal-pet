package handlers

import (
	"math/rand"
	"time"

	"github.com/kirkegaard/terminal-pet/pkg/pet"
)

func StartGame() (inGame bool, guessesLeft int, score int, number int, showResult bool, lastNumber int, lastGuessWasCorrect bool, animState string) {
	rand.Seed(time.Now().UnixNano())

	inGame = true
	guessesLeft = 5 // The player gets exactly 5 guesses total (5 rounds)
	score = 0
	number = rand.Intn(9) + 1 // 1-9
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
	newGameGuessesLeft := gameGuessesLeft - 1 // Decrement guesses left with each guess, regardless of outcome
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

	p.Weight -= 1
	if p.Weight < 10 { // Don't let pet get too thin
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
	return pet.NewPet(name, time.Now(), parent)
}
