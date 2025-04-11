package pet

import (
	"time"

	"github.com/kirkegaard/terminal-pet/pkg/pet/ascii"
)

type Pet struct {
	Name       string    `json:"name"`
	BirthDate  time.Time `json:"birthDate"`
	Parent     *Parent
	Hunger     int `json:"hunger"`
	Happiness  int `json:"happiness"`
	Discipline int
	Health     int       `json:"health"`
	Weight     int       `json:"weight"`
	IsSick     bool      `json:"isSick"`
	HasPooped  bool      `json:"hasPooped"`
	LightsOn   bool      `json:"lightsOn"`
	LastAction time.Time `json:"lastAction"`
	LastVisit  time.Time `json:"lastVisit"`
}

func NewPet(name string, birthday time.Time, parent *Parent) *Pet {
	return &Pet{
		Name:       name,
		BirthDate:  birthday,
		Parent:     parent,
		Hunger:     0,
		Happiness:  0,
		Discipline: 0,
		Health:     100,
		Weight:     1,
		IsSick:     false,
		HasPooped:  false,
		LightsOn:   true,
		LastAction: time.Now(),
		LastVisit:  time.Now(),
	}
}

func (p *Pet) String() string {
	return p.Name
}

// GetState returns the current state of the pet for animation purposes
func (p *Pet) GetState() ascii.PetState {
	if p.IsDead() {
		return ascii.StateDead
	}

	if !p.LightsOn {
		return ascii.StateSleeping
	}

	if p.IsSick {
		return ascii.StateSick
	}

	if p.Health < 50 {
		return ascii.StateSick
	}

	if p.Hunger > 70 {
		return ascii.StateHungry
	}

	if p.Happiness < 30 {
		return ascii.StateSad
	}

	if p.Happiness >= 80 {
		return ascii.StateHappy
	}

	return ascii.StateIdle
}

// Pet life stages
const (
	StageBaby   = "Baby"
	StageChild  = "Child"
	StageTeen   = "Teen"
	StageAdult  = "Adult"
	StageSenior = "Senior"
)

func (p *Pet) Age() int {
	return int(time.Since(p.BirthDate).Hours())
}

func (p *Pet) AgeInYears() int {
	return int(time.Since(p.BirthDate).Hours() / 24 / 15)
}

func (p *Pet) LifeStage() string {
	ageInYears := p.AgeInYears()

	switch {
	case ageInYears < 1:
		return StageBaby
	case ageInYears < 3:
		return StageChild
	case ageInYears < 6:
		return StageTeen
	case ageInYears < 12:
		return StageAdult
	default:
		return StageSenior
	}
}

func (p *Pet) IsAdult() bool {
	stage := p.LifeStage()
	return stage == StageAdult || stage == StageSenior
}

func (p *Pet) IsSenior() bool {
	return p.LifeStage() == StageSenior
}

func (p *Pet) IsDead() bool {
	return p.Health <= 0
}

func (p *Pet) Feed(food *Food) {
	p.LastAction = time.Now()

	if p.Hunger <= 0 {
		p.Health -= 2
		p.Weight += food.Weight
		return
	}

	p.Hunger -= food.Hunger
	if p.Hunger < 0 {
		p.Hunger = 0
	}

	p.Weight += food.Weight
	p.Health += food.Health

	if p.Health > 100 {
		p.Health = 100
	}
}

func (p *Pet) Play() {
	p.LastAction = time.Now()

	p.Happiness += 10
	if p.Happiness > 100 {
		p.Happiness = 100
	}

	p.Hunger += 5
	if p.Hunger > 100 {
		p.Hunger = 100
	}

	p.Weight -= 1
}

func (p *Pet) GiveMedicine() {
	p.LastAction = time.Now()

	if p.IsSick {
		p.IsSick = false
		p.Health += 20
		if p.Health > 100 {
			p.Health = 100
		}

		p.Happiness -= 5
		if p.Happiness < 0 {
			p.Happiness = 0
		}
	} else {
		p.Health -= 10
		if p.Health < 0 {
			p.Health = 0
		}

		p.Happiness -= 10
		if p.Happiness < 0 {
			p.Happiness = 0
		}
	}
}

func (p *Pet) Clean() {
	p.LastAction = time.Now()

	if p.HasPooped {
		p.HasPooped = false
		p.Happiness += 5
		if p.Happiness > 100 {
			p.Happiness = 100
		}

		p.Health += 5
		if p.Health > 100 {
			p.Health = 100
		}
	}
}

func (p *Pet) ToggleLights() {
	p.LastAction = time.Now()
	p.LightsOn = !p.LightsOn
}
