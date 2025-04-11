package pet

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

func Burger() (food *Food) {
	return NewFood("Burger", 1, 10, 30)
}

func Cake() (food *Food) {
	return NewFood("Cake", 10, 0, 20)
}
