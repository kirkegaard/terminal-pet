package models

import "time"

type Pet struct {
	ID         int       `db:"id"`
	Name       string    `db:"name"`
	BirthDate  time.Time `db:"birthday"`
	ParentID   int       `db:"parent_id"`
	Hunger     int       `db:"hunger"`
	Happiness  int       `db:"happiness"`
	Discipline int       `db:"discipline"`
	Health     int       `db:"health"`
	Weight     int       `db:"weight"`
	IsSick     bool      `db:"is_sick"`
	HasPooped  bool      `db:"has_pooped"`
	LightsOn   bool      `db:"lights_on"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
