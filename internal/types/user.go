package types

import "github.com/google/uuid"

type User struct {
	ID         uuid.UUID `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Img        *string   `db:"img" json:"img"`
	Email      string    `db:"email" json:"email"`
	Phone      string    `db:"phone" json:"phone"`
	Password   string    `db:"password" json:"-"`
	Created_at string    `db:"created_at" json:"created_at"`
	Updated_at string    `db:"updated_at" json:"updated_at"`
}
