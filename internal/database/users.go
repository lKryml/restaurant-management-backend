package database

import (
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"reflect"
	"restaurant-management-backend/internal/types"
	"strings"
)

var QB = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func (s *service) ListUsers() ([]types.User, error) {
	var users []types.User
	query, args, err := QB.Select("*").From("users").ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql query builder failed %w", err)
	}
	if err := s.db.Select(&users, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *service) GetUserByID(id string) (*types.User, error) {
	user := &types.User{}
	query, args, err := QB.Select("*").From("users").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql query builder failed %w", err)
	}
	if err := s.db.Get(user, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}
	return user, nil
}

func (s *service) CreateUser(user types.User) (uuid.UUID, error) {
	var id uuid.UUID
	query, args, err := InsertTypeSQL(user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error inserting user: %w", err)
	}

	fmt.Println(query, args)
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error executing insert query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return uuid.Nil, fmt.Errorf("error getting affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return uuid.Nil, fmt.Errorf("insert query did not affect any rows")
	}

	err = s.db.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error getting inserted ID: %w", err)
	}
	//_, err := s.db.NamedQuery("INSERT INTO users (name,email,password,phone) VALUES (:name,:email,:password,:phone) RETURNING id", user)
	//fmt.Println(id)

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

func InsertTypeSQL(data interface{}) (string, []interface{}, error) {
	v := reflect.ValueOf(data)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("data must be a struct")
	}

	// add s to type name user = users, vendor = vendors
	tableName := strings.ToLower(t.Name()) + "s"

	columns := []string{}
	values := []interface{}{}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			if !v.Field(i).IsZero() {
				columns = append(columns, dbTag)
				values = append(values, v.Field(i).Interface())
			}
		}
	}

	if len(columns) == 0 {
		return "", nil, fmt.Errorf("struct is empty")
	}

	insertBuilder := QB.Insert(tableName).
		Columns(columns...).
		Values(values...)

	return insertBuilder.ToSql()
}
