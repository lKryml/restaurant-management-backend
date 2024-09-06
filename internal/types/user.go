package types

type User struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Img        string `db:"img" json:"img"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}
