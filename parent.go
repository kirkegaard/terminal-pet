package main

type Parent struct {
	uuid string
}

func NewParent(uuid string) *Parent {
	return &Parent{
		uuid: uuid,
	}
}
