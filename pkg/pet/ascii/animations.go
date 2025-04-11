package ascii

// PetState represents the current state of the pet
type PetState string

// Pet states
const (
	StateDead     PetState = "dead"
	StateSleeping PetState = "sleeping"
	StateSick     PetState = "sick"
	StateHungry   PetState = "hungry"
	StateSad      PetState = "sad"
	StateHappy    PetState = "happy"
	StateIdle     PetState = "idle"
)

type Animation struct {
	Name   string
	Frames []string
	FPS    int
}

var (
	Happy = Animation{
		Name: "Happy",
		Frames: []string{
			`
 /\_/\
( ^.^ )
 > ^ <
`,
			`
 /\_/\
( ^o^ )
 > ^ <
`,
		},
		FPS: 2,
	}

	Idle = Animation{
		Name: "Idle",
		Frames: []string{
			`
 /\_/\
( ^.^ )
 > ^ <
`,
			`
 /\_/\
( ^-^ )
 > ^ <
`,
			`
 /\_/\
( ^.^ )
 > ^ <
`,
			`
 /\_/\
( -.- )
 > ^ <
`,
			`
 /\_/\
( ^.^ )
 > ^ <
`,
		},
		FPS: 1,
	}

	Sad = Animation{
		Name: "Sad",
		Frames: []string{
			`
 /\_/\
( T.T )
 > v <
`,
			`
 /\_/\
( u.u )
 > v <
`,
		},
		FPS: 2,
	}

	Sick = Animation{
		Name: "Sick",
		Frames: []string{
			`
 /\_/\
( @.@ )
 > x <
`,
			`
 /\_/\
( @-@ )
 > x <
`,
		},
		FPS: 2,
	}

	Hungry = Animation{
		Name: "Hungry",
		Frames: []string{
			`
 /\_/\
( o.o )
 > o <
`,
			`
 /\_/\
( o.o )
(> o <)
`,
		},
		FPS: 2,
	}

	Sleepy = Animation{
		Name: "Sleepy",
		Frames: []string{
			`
 /\_/\
(-.-  )
 > z <
`,
			`
 /\_/\
(-.-  )
 > Z <
`,
		},
		FPS: 2,
	}

	Dead = Animation{
		Name: "Dead",
		Frames: []string{
			`
 /\_/\
 ( x.x )
 > _ <
`,
		},
		FPS: 1,
	}

	Playing = Animation{
		Name: "Playing",
		Frames: []string{
			`
 /\_/\
( ^.^ )  ◯
 > ^ <   
`,
			`
 /\_/\
( ^o^ ) ◯
 > ^ <  
`,
			`
 /\_/\
( ^.^ )◯
 > ^ <   
`,
		},
		FPS: 3,
	}

	Eating = Animation{
		Name: "Eating",
		Frames: []string{
			`
 /\_/\
( o.o )    🍔
 > ^ <
`,
			`
 /\_/\
( o.o )  🍔
 > ^ <
`,
			`
 /\_/\
( o-o )🍔
 > ^ <
`,
			`
 /\_/\
( ^-^ )
 > ^ <
`,
		},
		FPS: 4,
	}

	CakeEating = Animation{
		Name: "CakeEating",
		Frames: []string{
			`
 /\_/\
( o.o )    🍰
 > ^ <
`,
			`
 /\_/\
( o.o )  🍰
 > ^ <
`,
			`
 /\_/\
( o-o )🍰
 > ^ <
`,
			`
 /\_/\
( ^-^ )
 > ^ <
`,
		},
		FPS: 4,
	}

	RightEyeBlink = `
 /\_/\
( -.^ )
 > ^ <
`

	LeftEyeBlink = `
 /\_/\
( ^.- )
 > ^ <
`

	LightsOff = Animation{
		Name: "Sleeping",
		Frames: []string{
			`
   z
  z
 Z
`,
			`
 z
  Z
   z
`,
		},
		FPS: 2,
	}
)

// GetAnimationForState returns the appropriate animation for the given pet state
func GetAnimationForState(state PetState) Animation {
	switch state {
	case StateDead:
		return Dead
	case StateSleeping:
		return LightsOff
	case StateSick:
		return Sick
	case StateHungry:
		return Hungry
	case StateSad:
		return Sad
	case StateHappy:
		return Happy
	default:
		return Idle
	}
}
