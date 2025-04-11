package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kirkegaard/terminal-pet/pkg/pet/ascii"
)

// RenderGameView renders the Higher or Lower game UI
func RenderGameView(
	baseView string,
	width int,
	currentFrame int,
	animState string,
	petPosition int,
	showResult bool,
	lastGuessWasCorrect bool,
	gameNumber int,
	lastNumber int,
	gameGuessesLeft int,
	gameScore int,
	inGame bool,
) string {
	var sb strings.Builder
	sb.WriteString(baseView)

	// Game title
	gameTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF6700")).
		Padding(0, 1).
		Render(" HIGHER OR LOWER GAME ")

	// Center the title
	padding := (width - lipgloss.Width(gameTitle)) / 2
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(gameTitle)
	sb.WriteString("\n\n")

	// Show pet animation with current number
	var animation ascii.Animation

	// Pick pet animation based on game state
	if showResult {
		if lastGuessWasCorrect {
			animation = ascii.Happy
		} else {
			animation = ascii.Sad
		}
	} else {
		animation = ascii.Playing
	}

	// Get current frame
	frameIdx := currentFrame
	if frameIdx >= len(animation.Frames) {
		frameIdx = 0
	}

	var frame string
	if len(animation.Frames) > 0 {
		frame = animation.Frames[frameIdx]
	} else {
		frame = ascii.Happy.Frames[0] // Default frame
	}

	// Add spacing
	sb.WriteString("\n")

	// Base padding
	basePadding := 10

	// Create a simple number display without a bubble
	var numberStyle lipgloss.Style

	// Style number based on result state
	if showResult {
		if lastGuessWasCorrect {
			numberStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00AA00"))
		} else {
			numberStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#AA0000"))
		}
	} else {
		numberStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#5F9EF3"))
	}

	// Make number bigger with double spacing
	numberDisplay := numberStyle.Render(fmt.Sprintf(" %d ", gameNumber))

	// Calculate the right position for the number
	rightPosForNumber := basePadding + petPosition + 10

	// Render each line of the frame with the number in the right position
	petLines := strings.Split(frame, "\n")
	for i, line := range petLines {
		actualPadding := basePadding + petPosition
		if actualPadding < 0 {
			actualPadding = 0
		}

		sb.WriteString(strings.Repeat(" ", actualPadding))
		sb.WriteString(line)

		// Add the number next to the second line of the pet (head level)
		if i == 1 && len(petLines) > 2 {
			extraSpacing := rightPosForNumber - actualPadding - len(line)
			if extraSpacing < 1 {
				extraSpacing = 1
			}
			sb.WriteString(strings.Repeat(" ", extraSpacing))
			sb.WriteString(numberDisplay)
		}

		sb.WriteString("\n")
	}

	// If showing result, display the previous number too
	if showResult {
		sb.WriteString("\n")

		// Center the previous number
		prevResult := fmt.Sprintf("Previous Number: %d", lastNumber)
		resultLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render(prevResult)

		padding = (width - lipgloss.Width(resultLine)) / 2
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(resultLine)
		sb.WriteString("\n")

		// Add result text
		var resultText string
		if lastGuessWasCorrect {
			resultText = "Correct! ✓"
			resultText = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00AA00")).
				Render(resultText)
		} else {
			resultText = "Wrong! ✗"
			resultText = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#AA0000")).
				Render(resultText)
		}

		padding = (width - lipgloss.Width(resultText)) / 2
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(resultText)
		sb.WriteString("\n\n")
	}

	// Game progress display right below the pet
	var scoreText string
	if gameGuessesLeft <= 0 {
		scoreText = fmt.Sprintf("GAME OVER! Final score: %d/5", gameScore)
	} else {
		scoreText = fmt.Sprintf("Score: %d/5   Guesses left: %d", gameScore, gameGuessesLeft)
	}

	scoreDisplay := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5F9EF3")).
		Bold(true).
		Render(scoreText)

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat(" ", basePadding))
	sb.WriteString(scoreDisplay)
	sb.WriteString("\n\n")

	// Draw game instructions
	if inGame {
		instructions := "← (Left/H): Lower   → (Right/L): Higher   ESC: Exit"
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat(" ", basePadding))
		sb.WriteString(instructions)
	}

	return sb.String()
}
