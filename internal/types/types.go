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
