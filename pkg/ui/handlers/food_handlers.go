package handlers

import (
	"github.com/kirkegaard/terminal-pet/pkg/pet"
)

func HandleFoodSubmenu(key string, cursor int, optionCount int) (keepSubmenu bool, newCursor int) {
	newCursor = cursor

	switch key {
	case "esc":
		return false, cursor
	case "left":
		if cursor > 0 {
			newCursor = cursor - 1
		}
	case "right":
		if cursor < optionCount-1 {
			newCursor = cursor + 1
		}
	}

	return true, newCursor
}

func HandleFoodSelection(key string, cursor int, optionCount int) (stayInFoodMode bool, newCursor int, selected bool) {
	newCursor = cursor

	switch key {
	case "esc":
		return false, cursor, false
	case "up", "down", "left", "right":
		if key == "up" || key == "left" {
			if cursor > 0 {
				newCursor = cursor - 1
			}
		} else if key == "down" || key == "right" {
			if cursor < optionCount-1 {
				newCursor = cursor + 1
			}
		}
	case "enter", " ":
		return false, cursor, true
	}

	return true, newCursor, false
}

func FeedPet(foodIndex int, petObj *pet.Pet) (string, *pet.Pet) {
	switch foodIndex {
	case 0: // Burger
		if petObj.Hunger <= 0 {
			return "idle", petObj
		}
		burger := pet.Burger()
		petObj.Feed(burger)
		petObj.Weight += 2
		return "eating", petObj
	case 1: // Cake
		cake := pet.Cake()
		petObj.Feed(cake)
		petObj.Weight += 5
		return "cakeEating", petObj
	}

	return "idle", petObj
}
