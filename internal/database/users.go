package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/logger"
	"restaurant-management-backend/internal/types"
)

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

	query, args, err := InsertBUILDER(user, fmt.Sprintf("RETURNING id,%s,created_at,updated_at", helpers.ImageFormat))
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
func (s *service) UpdateUser(newUser types.User, id string) (*types.User, error) {
	existingUser, err := s.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("error fetching existing user: %w", err)
	}

	var oldImage string
	if existingUser.Img != nil {
		oldImage = *existingUser.Img
	}

	query, args, err := UpdateBUILDER(newUser, id, "RETURNING *, "+helpers.ImageFormat)
	if err != nil {
		return nil, fmt.Errorf("error building update query: %w", err)
	}
	fmt.Println(query)
	var updatedUser types.User
	err = s.db.QueryRowx(query, args...).StructScan(&updatedUser)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}
	if newUser.Img != nil && *newUser.Img != oldImage {
		if oldImage != "" {
			if err = helpers.DeleteFile(oldImage); err != nil {
				logger.Log.WithError(err).Error("Failed to delete old user image")
			}
		}
	}

	return &updatedUser, nil
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

func (s *service) GetRoles(user *types.User) error {
	user.Roles = []int{}

	query, args, err := QB.Select("roles.id").
		From("roles").
		LeftJoin("user_roles ON user_roles.role_id = roles.id").
		LeftJoin("users ON user_roles.user_id = users.id").
		Where(squirrel.Eq{"users.id": user.ID}).
		OrderBy("roles.id").
		ToSql()
	if err != nil {
		return err
	}

	return s.db.SelectContext(context.Background(), &user.Roles, query, args...)
}
