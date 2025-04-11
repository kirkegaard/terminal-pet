package views

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Custom styles for the rename view
var (
	renameBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5F9EF3")).
			Padding(1, 3).
			Align(lipgloss.Center)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF00FF")).
			Padding(0, 1).
			Margin(1, 0)
)

// RenderRenameView renders the pet rename UI
func RenderRenameView(
	baseOutput string,
	width int,
	newName string,
) string {
	// Create the inner content first
	title := titleStyle.Render("✨ Rename Your Pet ✨")
	instruction := infoStyle.Render("Enter a new name for your pet:")

	// Input field with blinking cursor
	shouldShowCursor := time.Now().Second()%2 == 0
	cursor := "▌"
	if !shouldShowCursor {
		cursor = " "
	}

	styledInput := inputBoxStyle.Render(highlightStyle.Render(newName + cursor))
	controls := infoStyle.Render("Press Enter to confirm, ESC to cancel")

	// Join all components with proper spacing
	innerContent := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"", // Empty line for spacing
		instruction,
		styledInput,
		"", // Empty line for spacing
		controls,
	)

	// Calculate appropriate width for the box
	boxWidth := width - 20 // More margin to ensure box fits
	if boxWidth < 50 {
		boxWidth = 50 // Slightly wider minimum
	}

	// Apply box style to entire content
	boxedContent := renameBoxStyle.
		Width(boxWidth).
		Render(innerContent)

	// Center the box in the terminal
	centered := lipgloss.Place(
		width,
		10, // Height allocation
		lipgloss.Center,
		lipgloss.Center,
		boxedContent,
	)

	var output strings.Builder
	output.WriteString(baseOutput)
	output.WriteString(centered)

	return output.String()
}
