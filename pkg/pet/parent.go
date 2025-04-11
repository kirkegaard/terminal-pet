package pet

type Parent struct {
	ID   int
	Name string
}

func NewParent(id int, name string) *Parent {
	return &Parent{
		ID:   id,
		Name: name,
	}
}
