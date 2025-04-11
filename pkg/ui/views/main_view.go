package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
	"github.com/kirkegaard/terminal-pet/pkg/pet"
	"github.com/kirkegaard/terminal-pet/pkg/pet/ascii"
	"github.com/kirkegaard/terminal-pet/pkg/ui/keymap"
)

var choices = []string{"Feed", "Clean", "Play", "Medicine", "Rename", "Toggle Lights", "Quit"}

var (
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5F9EF3")).
			Bold(true).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5F9EF3")).
			Bold(true)

	disabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5F9EF3"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	foodSubmenuStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FF00FF")).
				Padding(0, 1)
)

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func GetPetState(pet *pet.Pet) string {
	if pet.IsDead() {
		return "Dead"
	}

	if !pet.LightsOn {
		return "Sleeping"
	}

	if pet.IsSick {
		return "Sick"
	}

	if pet.Hunger > 90 {
		return "Starving"
	} else if pet.Hunger > 70 {
		return "Hungry"
	}

	if pet.Happiness < 20 {
		return "Depressed"
	} else if pet.Happiness < 40 {
		return "Sad"
	} else if pet.Happiness >= 80 {
		return "Happy"
	}

	if pet.Health < 30 {
		return "Unhealthy"
	}

	if pet.HasPooped {
		return "Needs cleaning"
	}

	return "Content"
}

func RenderMainView(
	pet *pet.Pet,
	currentAnim ascii.Animation,
	currentFrame int,
	petPosition int,
	width int,
	showStats bool,
	showHelp bool,
	help help.Model,
	keys keymap.KeyMap,
	cursor int,
	selectedAction int,
	debugMode bool,
	showFoodSubmenu bool,
	foodSubmenuCursor int,
	foodOptions []string,
) string {
	var output strings.Builder

	basePadding := 5

	petName := fmt.Sprintf(" %s ", pet.Name)
	output.WriteString(titleStyle.Render(petName))
	output.WriteString("\n\n")

	frameStr := ""
	if len(currentAnim.Frames) > 0 && currentFrame < len(currentAnim.Frames) {
		frameStr = currentAnim.Frames[currentFrame]
	}

	lines := strings.Split(frameStr, "\n")
	offsetLines := make([]string, len(lines))

	for i, line := range lines {
		offsetLines[i] = strings.Repeat(" ", basePadding+petPosition) + line
	}

	if pet.IsSick || pet.HasPooped {
		if len(offsetLines) >= 2 {
			statusIndicators := ""
			if pet.IsSick {
				statusIndicators += "☠️"
			}
			if pet.HasPooped {
				statusIndicators += "💩"
			}
			offsetLines[1] = offsetLines[1] + " " + statusIndicators
		}
	}

	frameStr = strings.Join(offsetLines, "\n")

	for _, line := range strings.Split(frameStr, "\n") {
		output.WriteString(line)
		output.WriteString("\n")
	}

	output.WriteString("\n")

	if showStats {
		getHearts := func(percentage int) string {
			percentage = clamp(percentage, 0, 100)

			fullHearts := percentage / 20
			halfHeart := percentage%20 >= 10

			hearts := strings.Repeat("♥ ", fullHearts)
			if halfHeart {
				hearts += "♡ "
			}
			return hearts
		}

		ageDays := pet.Age()
		lifeStage := pet.LifeStage()
		petState := GetPetState(pet)

		stateLabel := infoStyle.Render("State:")
		stateValue := fmt.Sprintf(" %s", petState)
		output.WriteString(stateLabel + stateValue + "\n")

		ageLabel := infoStyle.Render("Age:")
		ageValue := fmt.Sprintf(" %d days (%s)", ageDays, lifeStage)
		output.WriteString(ageLabel + ageValue + "\n")

		healthLabel := infoStyle.Render("Health:")
		healthHearts := getHearts(pet.Health)
		output.WriteString(healthLabel + " " + healthHearts + "\n")

		hungerLabel := infoStyle.Render("Hunger:")
		hungerHearts := getHearts(100 - pet.Hunger)
		output.WriteString(hungerLabel + " " + hungerHearts + "\n")

		happinessLabel := infoStyle.Render("Happiness:")
		happinessHearts := getHearts(pet.Happiness)
		output.WriteString(happinessLabel + " " + happinessHearts + "\n")

		weightLabel := infoStyle.Render("Weight:")

		weightStr := fmt.Sprintf(" %d kg", pet.Weight)
		if pet.Weight > 100 {
			weightStr += " 🍔 (Obese)"
		} else if pet.Weight > 75 {
			weightStr += " 🍰 (Fat)"
		} else if pet.Weight > 50 {
			weightStr += " (Normal)"
		} else if pet.Weight > 25 {
			weightStr += " (Fit)"
		} else {
			weightStr += " 🏃 (Athletic)"
		}

		output.WriteString(weightLabel + weightStr + "\n")
	}

	output.WriteString("\n\n")

	for i, choice := range choices {
		// 5=Toggle Lights, 6=Quit
		if !pet.LightsOn && i != 5 && i != 6 {
			output.WriteString(disabledStyle.Render(" " + choice + " "))
		} else if i == cursor {
			if i == selectedAction {
				output.WriteString(selectedStyle.Render(choice))
			} else {
				output.WriteString(highlightStyle.Render(choice))
			}
		} else {
			if i == selectedAction {
				output.WriteString(selectedStyle.Render(choice))
			} else {
				output.WriteString(normalStyle.Render(" " + choice + " "))
			}
		}
		output.WriteString(" ")
	}
	output.WriteString("\n")

	if !pet.LightsOn || pet.IsSick || pet.HasPooped {
		output.WriteString("\n")

		if !pet.LightsOn {
			sleepMsg := lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("#AAAAAA")).
				Render("Your pet is sleeping. Toggle the lights to wake them up!")
			output.WriteString(sleepMsg)
			output.WriteString("\n")
		}

		if pet.IsSick {
			sickMsg := warningStyle.Render("Your pet is sick! 🤒 Give it medicine!")
			output.WriteString(sickMsg)
			output.WriteString("\n")
		}

		if pet.HasPooped {
			poopMsg := warningStyle.Render("Your pet needs cleaning! 💩")
			output.WriteString(poopMsg)
			output.WriteString("\n")
		}
	}

	if showFoodSubmenu {
		output.WriteString("\n")
		output.WriteString(strings.Repeat(" ", basePadding+10))

		output.WriteString(infoStyle.Render("Select food:") + " ")

		for i, option := range foodOptions {
			if i == foodSubmenuCursor {
				output.WriteString(highlightStyle.Render(option))
			} else {
				output.WriteString(normalStyle.Render(" " + option + " "))
			}
			output.WriteString(" ")
		}

		output.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Render(" (ESC to cancel, ←→ to navigate)"))
	}

	if pet.IsDead() {
		output.WriteString("\n\n")
		deathMsg := warningStyle.Render("Your pet has died! 😢")
		output.WriteString(deathMsg)
	}

	// Display key help at the bottom
	if showHelp {
		output.WriteString("\n\n")
		output.WriteString(help.View(keys))
	}

	return output.String()
}
