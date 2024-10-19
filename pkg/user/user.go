package main

import (
	"time"
)

type User struct {
	uuid      string
	localTime time.Time
}

func NewUser(uuid string, localTime time.Time) *User {
	return &User{
		uuid:      uuid,
		localTime: localTime,
	}
}
