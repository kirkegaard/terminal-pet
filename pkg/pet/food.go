package main

type Food struct {
	Name   string
	Weight int
	Health int
	Hunger int
}

func NewFood(name string, weight int, health int, hunger int) *Food {
	return &Food{
		Name:   name,
		Weight: weight,
		Health: health,
		Hunger: hunger,
	}
}

func Burger() (cake *Food) {
	return NewFood("Burger", 1, 10, 10)
}

func Cake() (cake *Food) {
	return NewFood("Cake", 10, 0, 0)
}
