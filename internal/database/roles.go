package database

import (
	"database/sql"
	"net/url"
	"restaurant-management-backend/internal/types"
)

func (s *service) FetchRoles(queryParams map[string][]string) ([]types.Role, *types.Meta, error) {
	var roles []types.Role

	urlValues := url.Values{}
	for key, values := range queryParams {
		for _, value := range values {
			urlValues.Add(key, value)
		}
	}

	columns := []string{"id", "name"}

	searchColumns := []string{"name"}

	meta, err := s.BuildQuery(
		&roles,
		"roles",
		[]string{},
		columns,
		searchColumns,
		urlValues,
		[]string{},
	)

	if err != nil {
		return nil, nil, err
	}

	if roles == nil {
		roles = []types.Role{}
	}

	return roles, meta, nil
}

func (s *service) FetchRole(id string) (types.Role, error) {
	var role types.Role
	query, args, err := QB.Select("*").From("roles").Where("id = ?", id).ToSql()
	if err != nil {
		return role, err
	}
	err = s.db.Get(&role, query, args...)
	return role, err
}

func (s *service) VerifyRoleExists(roleID string) error {
	var role []types.Role
	query, args, err := QB.Select("*").From("roles").Where("id = ?", roleID).ToSql()
	if err != nil {
		return err
	}
	if err := s.db.Select(&role, query, args...); err != nil {
		return err
	}
	if len(role) == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *service) GrantRole(userID, roleID string) error {
	query, args, err := QB.Insert("user_roles").Columns("user_id", "role_id").Values(userID, roleID).ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}

func (s *service) RevokeRole(userID, roleID string) (int64, error) {
	query, args, err := QB.Delete("user_roles").Where("user_id = ? AND role_id = ?", userID, roleID).ToSql()
	if err != nil {
		return 0, err
	}
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
