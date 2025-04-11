package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kirkegaard/terminal-pet/pkg/pet"
	"github.com/kirkegaard/terminal-pet/pkg/pet/ascii"
)

// RenderGameOver renders the game over screen
func RenderGameOver(
	baseView string,
	width int,
	pet *pet.Pet,
	gameOverCursor int,
) string {
	var sb strings.Builder
	sb.WriteString(baseView)

	// Game over title
	gameOverTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF0000")).
		Padding(0, 1).
		Render(" GAME OVER ")

	// Center the title
	padding := (width - lipgloss.Width(gameOverTitle)) / 2
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(gameOverTitle)
	sb.WriteString("\n\n")

	// Pet animation (dead)
	frame := ascii.Dead.Frames[0]

	// Center the frame
	for _, line := range strings.Split(frame, "\n") {
		padding = (width - len(line)) / 2
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Death message
	deathMsg := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF0000")).
		Render("Your pet has died! 😢")

	padding = (width - lipgloss.Width(deathMsg)) / 2
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(deathMsg)
	sb.WriteString("\n\n")

	// Pet stats
	ageDays := pet.Age()
	lifeStage := pet.LifeStage()
	petAge := fmt.Sprintf("Age: %d days (%s)", ageDays, lifeStage)

	padding = (width - lipgloss.Width(petAge)) / 2
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(petAge)
	sb.WriteString("\n\n")

	// Options
	options := []string{"Restart", "Quit"}

	for i, option := range options {
		var cursor string
		var style lipgloss.Style

		if gameOverCursor == i {
			cursor = ">"
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#FF0000")).
				Bold(true)
		} else {
			cursor = " "
			style = lipgloss.NewStyle()
		}

		padding = (width - len(option) - 3) / 2
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", cursor, style.Render(option)))
	}

	return sb.String()
}
