package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
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
		logrus.WithError(err).Error("Failed to build SQL query")
		return nil, fmt.Errorf("internal server error")
	}
	if err := s.db.Get(user, query, args...); err != nil {
		if err == sql.ErrNoRows {
			logrus.WithField("id", id).Info("User not found")
			return nil, fmt.Errorf("user not found")
		}
		logrus.WithError(err).WithField("id", id).Error("Failed to fetch user from database")
		return nil, fmt.Errorf("internal server error")
	}
	return user, nil
}

func (s *service) CreateUser(user types.User) (*types.User, error) {
	//var (
	//	id         uuid.UUID
	//	created_at time.Time
	//	updated_at time.Time
	//)
	query, args, err := InsertTypeSQL(user)
	if err != nil {
		return nil, fmt.Errorf("error inserting user: %w", err)
	}
	err = s.db.QueryRowx(query, args...).StructScan(&user)
	if err != nil {
		//if err != nil {
		//	return uuid.Nil, fmt.Errorf("error executing insert query: %w", err)
		//}

		//double check for existing if bypassed the first one from query builder
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, fmt.Errorf("user with this email already exists")
		}

		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error inserting user: insert did not return an id!!! lol")
		}

		return nil, fmt.Errorf("error inserting user: %w", err)
	}

	//if id == uuid.Nil {
	//	return nil, fmt.Errorf("received nil UUID after insert")
	//}

	//rowsAffected, err := result.RowsAffected()
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("error getting affected rows: %w", err)
	//}
	//if rowsAffected == 0 {
	//	return uuid.Nil, fmt.Errorf("insert query did not affect any rows")
	//}
	//
	//err = s.db.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&id)
	//if err != nil {
	//	return uuid.Nil, fmt.Errorf("error getting inserted ID: %w", err)
	//}

	//createdUser := &types.User{
	//	ID:         id,
	//	Name:       user.Name,
	//	Img:        user.Img,
	//	Email:      user.Email,
	//	Phone:      user.Phone,
	//	Created_at: created_at.Format(time.RFC3339),
	//	Updated_at: updated_at.Format(time.RFC3339),
	//}

	return &user, nil
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
		Values(values...).
		Suffix("RETURNING *")

	return insertBuilder.ToSql()
}
