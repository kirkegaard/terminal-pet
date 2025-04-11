package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kirkegaard/terminal-pet/pkg/pet"
)

// GetDebugMenuItemCount returns the number of items in the debug menu
func GetDebugMenuItemCount() int {
	return 15 // Number of debug options
}

// RenderDebugMenu renders the debug menu UI
func RenderDebugMenu(
	baseView string,
	width int,
	pet *pet.Pet,
	debugCursor int,
) string {
	var sb strings.Builder
	sb.WriteString(baseView)

	// Debug menu title
	debugTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#CC0000")).
		Padding(0, 1).
		Render(" DEBUG MENU ")

	// Center the title
	padding := (width - lipgloss.Width(debugTitle)) / 2
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(debugTitle)
	sb.WriteString("\n\n")

	// Current pet state info
	stateInfo := fmt.Sprintf("Current State - Health: %d, Hunger: %d, Happiness: %d",
		pet.Health, pet.Hunger, pet.Happiness)

	stateDisplay := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Render(stateInfo)

	padding = (width - lipgloss.Width(stateDisplay)) / 2
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(stateDisplay)
	sb.WriteString("\n\n")

	// Debug options
	debugOptions := []string{
		"Toggle Sick",
		"Toggle Poop",
		"Toggle Dead",
		"Set Full Health (100)",
		"Set Low Health (30)",
		"Set Not Hungry (0)",
		"Set Very Hungry (90)",
		"Set Happy (100)",
		"Set Sad (10)",
		"Critical State (H=1, Hu=100, Ha=0)",
		"Reset All Stats",
		"Exit Debug Mode",
		"Set Obese (Weight=110)",
		"Set Normal Weight (Weight=50)",
		"Set Athletic (Weight=20)",
	}

	for i, option := range debugOptions {
		var cursor string
		var style lipgloss.Style

		if debugCursor == i {
			cursor = ">"
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#CC0000")).
				Bold(true)
		} else {
			cursor = " "
			style = lipgloss.NewStyle()
		}

		// Center the menu items
		padding = (width - len(option) - 3) / 2
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", cursor, style.Render(option)))
	}

	// Instructions
	sb.WriteString("\n")
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Render("Use arrow keys to navigate, Enter to select, ESC to cancel")

	padding = (width - lipgloss.Width(instructions)) / 2
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(instructions)

	return sb.String()
}
