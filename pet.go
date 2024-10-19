package main

import (
	"time"
)

type Pet struct {
	Name     string
	Birthday time.Time
	Parent   *Parent
}

func NewPet(name string, birthday time.Time, parent *Parent) *Pet {
	return &Pet{
		Name:     name,
		Birthday: birthday,
		Parent:   parent,
	}
}
