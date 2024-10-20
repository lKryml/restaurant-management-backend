package database

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"restaurant-management-backend/internal/logger"
	"restaurant-management-backend/internal/types"
)

func (s *service) GetUserByEmail(email string) (*types.User, error) {
	user := &types.User{}
	query, args, err := QB.Select("*").From("users").Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		logger.Log.WithError(err).Error("Failed to build SQL query")
		return nil, fmt.Errorf("internal server error %w", err)
	}
	if err := s.db.Get(user, query, args...); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) GrantDefaultRole(userID uuid.UUID) error {
	query, args, err := QB.Insert("user_roles").Columns("user_id", "role_id").Values(userID, 3).ToSql()
	if err != nil {
		return fmt.Errorf("error generate query: %w", err)
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error granting role: %w", err)
	}

	return nil
}
