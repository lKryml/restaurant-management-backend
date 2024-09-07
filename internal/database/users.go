package database

import (
	"fmt"
	"restaurant-management-backend/internal/types"
)

func (s *service) ListUsers() ([]types.User, error) {
	var users []types.User
	err := s.db.Select(&users, "SELECT * FROM users")
	return users, err
}

func (s *service) GetUserByID(id int) (types.User, error) {
	var user types.User

	err := s.db.Get(&user, "SELECT id, name FROM users WHERE id = $1", id)
	return user, err
}

func (s *service) CreateUser(user types.User) (int, error) {
	var id int
	_, err := s.db.NamedQuery("INSERT INTO users (name,email,password,phone) VALUES (:name,:email,:password,:phone) RETURNING id", user)
	fmt.Println(id)
	return id, err
}

func (s *service) UpdateUser(user types.User) error {
	_, err := s.db.Exec("UPDATE users SET name = $1 WHERE id = $2", user.Name, user.ID)
	return err
}

func (s *service) DeleteUser(id int) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}
