package types

import "github.com/google/uuid"

type User struct {
	ID         uuid.UUID `db:"id" json:"id,omitempty"`
	Name       string    `db:"name" json:"name,omitempty"`
	Img        *string   `db:"img" json:"img,omitempty"`
	Email      string    `db:"email" json:"email,omitempty"`
	Phone      string    `db:"phone" json:"phone,omitempty"`
	Password   string    `db:"password" json:"-"`
	Created_at string    `db:"created_at" json:"created_at,omitempty"`
	Updated_at string    `db:"updated_at" json:"updated_at,omitempty"`
}

type Meta struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	FirstPage   int `json:"first_page"`
	LastPage    int `json:"last_page"`
	From        int `json:"from"`
	To          int `json:"to"`
}
