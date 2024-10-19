package models

type User struct {
	ID        int    `db:"id"`
	Name      string `db:"name"`
	PublicKey string `db:"public_key"`
}
