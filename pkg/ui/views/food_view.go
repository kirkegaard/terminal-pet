package views

import (
	"fmt"
	"strings"

	"github.com/kirkegaard/terminal-pet/pkg/pet"
)

// RenderFoodSelectionView renders the food selection UI
func RenderFoodSelectionView(
	baseOutput string,
	width int,
	pet *pet.Pet,
	foodCursor int,
	foodOptions []string,
) string {
	var output strings.Builder
	output.WriteString(baseOutput)

	// Food selection title
	foodTitle := titleStyle.Render("🍔 Select Food 🍰")
	titleWidth := len(foodTitle) - 10 // Approximate adjustment for style codes
	titlePadding := (width - titleWidth) / 2
	if titlePadding < 0 {
		titlePadding = 0
	}

	output.WriteString(strings.Repeat(" ", titlePadding))
	output.WriteString(foodTitle)
	output.WriteString("\n\n")

	// Pet hunger status
	hungerStatus := fmt.Sprintf("Current Hunger: %d%%", pet.Hunger)
	hungerWidth := len(hungerStatus)
	hungerPadding := (width - hungerWidth) / 2
	if hungerPadding < 0 {
		hungerPadding = 0
	}

	output.WriteString(strings.Repeat(" ", hungerPadding))

	// Different colors based on hunger level
	if pet.Hunger < 30 {
		output.WriteString(infoStyle.Render(hungerStatus + " (Not very hungry)"))
	} else if pet.Hunger < 70 {
		output.WriteString(infoStyle.Render(hungerStatus + " (Hungry)"))
	} else {
		output.WriteString(warningStyle.Render(hungerStatus + " (Starving!)"))
	}

	output.WriteString("\n\n")

	// Food options with descriptions
	foodDescriptions := []string{
		"Burger - Nutritious meal (requires hunger > 20)",
		"Cake - Sweet treat (can feed anytime)",
	}

	for i, description := range foodDescriptions {
		// Highlight selected option
		if i == foodCursor {
			styledOption := highlightStyle.Render("› " + description + " ‹")
			optionWidth := len(styledOption) - 15 // Approximate adjustment for style codes
			optionPadding := (width - optionWidth) / 2
			if optionPadding < 0 {
				optionPadding = 0
			}

			output.WriteString(strings.Repeat(" ", optionPadding))
			output.WriteString(styledOption)
		} else {
			styledOption := normalStyle.Render("  " + description + "  ")
			optionWidth := len(styledOption) - 10 // Approximate adjustment for style codes
			optionPadding := (width - optionWidth) / 2
			if optionPadding < 0 {
				optionPadding = 0
			}

			output.WriteString(strings.Repeat(" ", optionPadding))
			output.WriteString(styledOption)
		}

		output.WriteString("\n")
	}

	output.WriteString("\n")

	// Food effects info
	burgerInfo := "Burger: +10 Health, -10 Hunger, +2 Weight"
	cakeInfo := "Cake: +0 Health, +0 Hunger, +5 Weight"

	infoWidth := len(burgerInfo)
	infoPadding := (width - infoWidth) / 2
	if infoPadding < 0 {
		infoPadding = 0
	}

	output.WriteString(strings.Repeat(" ", infoPadding))
	output.WriteString(infoStyle.Render(burgerInfo))
	output.WriteString("\n")

	output.WriteString(strings.Repeat(" ", infoPadding))
	output.WriteString(infoStyle.Render(cakeInfo))
	output.WriteString("\n\n")

	// Controls
	controls := "Arrow keys: Navigate   Enter: Select   ESC: Cancel"
	controlsWidth := len(controls)
	controlsPadding := (width - controlsWidth) / 2
	if controlsPadding < 0 {
		controlsPadding = 0
	}

	output.WriteString(strings.Repeat(" ", controlsPadding))
	output.WriteString(controls)

	return output.String()
}
