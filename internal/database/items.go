package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/logger"
	"restaurant-management-backend/internal/types"
	"strings"
	"time"
)

var itemColumns = []string{
	"id",
	"vendor_id",
	"name",
	"price",
	"created_at",
	"updated_at",
	helpers.ImageFormat,
}

func (s *service) ListItems(query map[string][]string) ([]types.Item, *types.Meta, error) {
	var items []types.Item

	urlValues := make(url.Values)
	for key, values := range query {
		for _, value := range values {
			urlValues.Add(key, value)
		}
	}

	columns := []string{
		"id",
		"vendor_id",
		"name",
		"price",
		"created_at",
		"updated_at",
		helpers.ImageFormat,
	}

	searchColumns := []string{"name", "price"}

	meta, err := s.BuildQuery(
		&items,
		"items",
		[]string{},
		columns,
		searchColumns,
		urlValues,
		[]string{},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to list items: %w", err)
	}

	if items == nil {
		items = []types.Item{}
	}

	return items, meta, nil
}

func (s *service) CreateItem(item types.Item, r *http.Request) (*types.Item, error) {
	item.ID = uuid.New()
	item.Created_at = time.Now()
	item.Updated_at = time.Now()

	if item.VendorId == uuid.Nil || item.Price == 0 || item.Name == "" {
		return nil, errors.New("missing required parameters")
	}

	img, err := helpers.HandleFileUpload(r, "items")
	if err != nil {
		return nil, err
	}
	item.Img = img

	query, args, err := QB.
		Insert("items").
		Columns("id", "vendor_id", "name", "price", "created_at", "updated_at", "img").
		Values(item.ID, item.VendorId, item.Name, item.Price, item.Created_at, item.Updated_at, item.Img).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(itemColumns, ", "))).
		ToSql()
	if err != nil {
		if item.Img != nil {
			helpers.DeleteFile(*item.Img)
		}
		return nil, fmt.Errorf("error generating query: %w", err)
	}

	err = s.db.QueryRowx(query, args...).StructScan(&item)
	if err != nil {
		if item.Img != nil {
			helpers.DeleteFile(*item.Img)
		}
		return nil, fmt.Errorf("error inserting item: %w", err)
	}

	return &item, nil
}

func (s *service) GetItemByID(id string) (*types.Item, error) {
	var item types.Item
	query, args, err := QB.Select(strings.Join(itemColumns, ", ")).
		From("items").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}
	if err := s.db.Get(&item, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item not found: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch item: %w", err)
	}
	return &item, nil
}

func (s *service) DeleteItem(id string) error {
	query, args, err := QB.Delete("items").Where(squirrel.Eq{"id": id}).Suffix("RETURNING img").ToSql()
	if err != nil {
		return fmt.Errorf("error building delete query: %w", err)
	}

	var img *string
	err = s.db.QueryRow(query, args...).Scan(&img)
	if err != nil {
		return fmt.Errorf("error deleting item: %w", err)
	}

	if img != nil {
		if err := helpers.DeleteFile(*img); err != nil {
			logger.Log.WithError(err).Error("Failed to delete item image")
		}
	}

	return nil
}

func (s *service) UpdateItem(id string, updates map[string]interface{}, r *http.Request) (*types.Item, error) {
	item, err := s.GetItemByID(id)
	if err != nil {
		return nil, err
	}

	var oldImg *string
	if item.Img != nil {
		oldImg = item.Img
	}

	img, err := helpers.HandleFileUpload(r, "items")
	if err != nil {
		return nil, err
	}
	updates["img"] = img

	updates["updated_at"] = time.Now()

	query, args, err := QB.Update("items").
		SetMap(updates).
		Where(squirrel.Eq{"id": id}).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(itemColumns, ", "))).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building update query: %w", err)
	}

	var updatedItem types.Item
	err = s.db.QueryRowx(query, args...).StructScan(&updatedItem)
	if err != nil {
		return nil, fmt.Errorf("error updating item: %w", err)
	}

	if oldImg != nil && img != nil {
		if err := helpers.DeleteFile(*oldImg); err != nil {
			logger.Log.WithError(err).Error("Failed to delete old item image")
		}
	}

	return &updatedItem, nil
}
