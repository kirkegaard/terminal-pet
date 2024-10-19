package main

import (
	"time"
)

type Pet struct {
	Name       string
	Birthday   time.Time
	Parent     *Parent
	Hunger     int
	Happiness  int
	Discipline int
	Health     int
	Weight     int
}

func NewPet(name string, birthday time.Time, parent *Parent) *Pet {
	return &Pet{
		Name:       name,
		Birthday:   birthday,
		Parent:     parent,
		Hunger:     0,
		Happiness:  0,
		Discipline: 0,
		Health:     100,
		Weight:     0,
	}
}

func (p *Pet) String() string {
	return p.Name
}

func (p *Pet) Age() int {
	return int(time.Since(p.Birthday).Hours() / 24 / 365)
}

func (p *Pet) IsAdult() bool {
	return p.Age() >= 3
}

func (p *Pet) IsSenior() bool {
	return p.Age() >= 7
}

func (p *Pet) IsDead() bool {
	return p.Age() >= 15 || p.Health <= 0
}

func (p *Pet) Feed(food *Food) {
	p.Hunger -= food.Hunger
	p.Weight += food.Weight
	p.Health += food.Health
}

func (p *Pet) Play() {
	p.Happiness += 10
	p.Hunger += 5
	p.Weight -= 1
}
