package database

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"restaurant-management-backend/internal/types"
)

var QB = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func (s *service) ListUsers() ([]types.User, error) {
	var users []types.User
	query, args, err := QB.Select("*").From("users").ToSql()
	if err != nil {
		fmt.Errorf("sql query builder failed %w", err)
	}
	if err := s.db.Select(&users, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *service) GetUserByID(id int) (types.User, error) {
	var user types.User

	err := s.db.Get(&user, "SELECT id, name FROM users WHERE id = $1", id)
	return user, err
}

func (s *service) CreateUser(user types.User) error {
	var id string
	_, err := s.db.NamedQuery("INSERT INTO users (name,email,password,phone) VALUES (:name,:email,:password,:phone) RETURNING id", user)
	fmt.Println(id)
	return err
}

func (s *service) UpdateUser(user types.User) error {
	_, err := s.db.Exec("UPDATE users SET name = $1 WHERE id = $2", user.Name, user.ID)
	return err
}

func (s *service) DeleteUser(id int) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}
