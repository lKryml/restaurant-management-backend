package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"reflect"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/logger"
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
		logger.Log.WithError(err).Error("Failed to build SQL query")
		return nil, fmt.Errorf("internal server error %w", err)
	}
	if err := s.db.Get(user, query, args...); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.WithField("id", id).Info("User not found")
			return nil, fmt.Errorf("user not found %w", err)
		}
		logger.Log.WithError(err).WithField("id", id).Error("Failed to fetch user from database")
		return nil, fmt.Errorf("internal server error %w", err)
	}
	return user, nil
}

func (s *service) CreateUser(user types.User) (*types.User, error) {
	imgFRM := fmt.Sprintf("CASE WHEN NULLIF(img,'') IS NOT NULL THEN FORMAT ('%s/%%s',img) ELSE NULL END AS img ", helpers.Domain)

	query, args, err := InsertTypeSQL(user, fmt.Sprintf("RETURNING id,%s,created_at,updated_at", imgFRM))
	if err != nil {
		return nil, fmt.Errorf("error inserting user: %w", err)
	}

	err = s.db.QueryRowx(query, args...).StructScan(&user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error inserting user: insert did not return anything!!! lol %w", err)
		}
		return nil, fmt.Errorf("error inserting user: %w", err)
	}
	if user.ID == uuid.Nil {
		return nil, fmt.Errorf("inserted user sucessfully, but failed to fetch id")
	}
	return &user, nil
}

func (s *service) UpdateUser(user types.User) error {
	_, err := s.db.Exec("UPDATE users SET name = $1 WHERE id = $2", user.Name, user.ID)
	return err
}

func (s *service) DeleteUser(id string) error {
	result, err := deleteById(s, id, "users", "RETURNING img")
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	if result != nil {
		err = helpers.DeleteFile(*result)
		if err != nil {
			return fmt.Errorf("error deleting user: %w", err)
		}
	}
	return nil
}

func deleteById(s *service, id string, table string, suffix ...string) (*string, error) {

	var data *string
	var deleteBuilder squirrel.DeleteBuilder
	if len(suffix) > 0 {
		deleteBuilder = QB.Delete(table).Where(squirrel.Eq{"id": id}).Suffix(suffix[0])
	} else {
		deleteBuilder = QB.Delete(table).Where(squirrel.Eq{"id": id})
	}
	query, args, err := deleteBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error deleting user failed building sql query: %w", err)
	}
	if len(suffix) > 0 {
		err = s.db.QueryRowx(query, args...).Scan(&data)
		if err != nil {
			return nil, fmt.Errorf("error deleting user failed query: %w", err)
		}
		return data, nil

	} else {
		result, err := s.db.Exec(query, args...)
		if err != nil {
			return nil, fmt.Errorf("error deleting user failed sql exec: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("error deleting user failed to fetch rows affected:%w", err)
		}
		if affected == 0 {
			return nil, fmt.Errorf("error deleting user: no rows affected")
		}

	}
	return nil, nil
}

func InsertTypeSQL(data interface{}, suffix ...string) (string, []interface{}, error) {
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

	var insertBuilder squirrel.InsertBuilder
	if len(suffix) > 0 {
		insertBuilder = QB.Insert(tableName).
			Columns(columns...).
			Values(values...).
			Suffix(suffix[0])
	} else {
		insertBuilder = QB.Insert(tableName).
			Columns(columns...).
			Values(values...)
	}

	return insertBuilder.ToSql()
}
