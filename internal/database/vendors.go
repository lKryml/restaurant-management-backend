package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"net/url"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/logger"
	"restaurant-management-backend/internal/types"
	"strings"
	"time"
)

var (
	vendorColumns = []string{
		"id",
		"name",
		"description",
		"created_at",
		"updated_at",
		helpers.ImageFormat,
	}
)

func (s *service) ListVendors(queryParams url.Values) ([]types.Vendor, *types.Meta, error) {
	var vendors []types.Vendor

	meta, err := s.BuildQuery(
		&vendors,
		"vendors",
		[]string{},
		vendorColumns,
		[]string{"name", "description"},
		queryParams,
		[]string{},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list vendors: %w", err)
	}

	if vendors == nil {
		vendors = []types.Vendor{}
	}

	return vendors, meta, nil
}

func (s *service) GetVendorByID(id string) (*types.Vendor, error) {
	vendor := &types.Vendor{}
	query, args, err := QB.Select(strings.Join(vendorColumns, ", ")).
		From("vendors").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		logger.Log.WithError(err).Error("Failed to build SQL query")
		return nil, fmt.Errorf("internal server error %w", err)
	}
	if err := s.db.Get(vendor, query, args...); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.WithField("id", id).Info("Vendor not found")
			return nil, fmt.Errorf("vendor not found %w", err)
		}
		logger.Log.WithError(err).WithField("id", id).Error("Failed to fetch vendor from database")
		return nil, fmt.Errorf("internal server error %w", err)
	}
	return vendor, nil
}

func (s *service) CreateVendor(vendor types.Vendor) (*types.Vendor, error) {
	vendor.ID = uuid.New()

	if vendor.Img != nil {
		*vendor.Img = strings.TrimPrefix(*vendor.Img, helpers.Domain+"/")
	}

	query, args, err := QB.
		Insert("vendors").
		Columns("id", "img", "name", "description").
		Values(vendor.ID, vendor.Img, vendor.Name, vendor.Description).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(vendorColumns, ", "))).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error inserting vendor: %w", err)
	}

	err = s.db.QueryRowx(query, args...).StructScan(&vendor)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error inserting vendor: insert did not return anything %w", err)
		}
		return nil, fmt.Errorf("error inserting vendor: %w", err)
	}

	return &vendor, nil
}

func (s *service) UpdateVendor(newVendor types.Vendor, id string) (*types.Vendor, error) {
	existingVendor, err := s.GetVendorByID(id)
	if err != nil {
		return nil, fmt.Errorf("error fetching existing vendor: %w", err)
	}

	var oldImage string
	if existingVendor.Img != nil {
		oldImage = *existingVendor.Img
	}

	if newVendor.Img != nil {
		*newVendor.Img = strings.TrimPrefix(*newVendor.Img, helpers.Domain+"/")
	}

	query, args, err := QB.
		Update("vendors").
		Set("img", newVendor.Img).
		Set("name", newVendor.Name).
		Set("description", newVendor.Description).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": id}).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(vendorColumns, ", "))).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building update query: %w", err)
	}

	var updatedVendor types.Vendor
	err = s.db.QueryRowx(query, args...).StructScan(&updatedVendor)
	if err != nil {
		return nil, fmt.Errorf("error updating vendor: %w", err)
	}

	if newVendor.Img != nil && *newVendor.Img != oldImage {
		if oldImage != "" {
			if err = helpers.DeleteFile(oldImage); err != nil {
				logger.Log.WithError(err).Error("Failed to delete old vendor image")
			}
		}
	}

	return &updatedVendor, nil
}

func (s *service) DeleteVendor(id string) error {
	query, args, err := QB.Delete("vendors").
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING img").
		ToSql()
	if err != nil {
		return fmt.Errorf("error building delete query: %w", err)
	}

	var img *string
	err = s.db.QueryRow(query, args...).Scan(&img)
	if err != nil {
		return fmt.Errorf("error deleting vendor: %w", err)
	}

	if img != nil {
		err = helpers.DeleteFile(*img)
		if err != nil {
			logger.Log.WithError(err).Error("Failed to delete vendor image")
		}
	}

	return nil
}

func (s *service) GrantAdmin(userID, vendorID string) error {
	query, args, err := QB.Insert("vendor_admins").
		Columns("user_id", "vendor_id").
		Values(userID, vendorID).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building grant admin query: %w", err)
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error granting admin: %w", err)
	}

	return nil
}

func (s *service) RevokeAdmin(userID, vendorID string) error {
	query, args, err := QB.Delete("vendor_admins").
		Where(squirrel.Eq{"user_id": userID, "vendor_id": vendorID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building revoke admin query: %w", err)
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error revoking admin: %w", err)
	}

	return nil
}

func (s *service) ListVendorAdmins(vendorID string) ([]types.User, error) {
	query, args, err := QB.Select("users.*").
		From("users").
		Join("vendor_admins ON vendor_admins.user_id = users.id").
		Where(squirrel.Eq{"vendor_admins.vendor_id": vendorID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building vendor admins query: %w", err)
	}

	var users []types.User
	err = s.db.Select(&users, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listing vendor admins: %w", err)
	}

	for i := range users {
		if err := s.GetRoles(&users[i]); err != nil {
			logger.Log.WithError(err).Error("Failed to get roles for user")
		}
	}

	if users == nil {
		users = []types.User{}
	}

	return users, nil
}
