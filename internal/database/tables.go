package database

import (
	"github.com/google/uuid"
	"net/http"
	"net/url"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/types"
)

func (s *service) FetchTables(queryParams url.Values) ([]types.Table, *types.Meta, error) {
	var tables []types.Table

	columns := []string{"id", "name", "vendor_id", "customer_id", "is_available", "is_needs_service"}

	searchColumns := []string{"name"}

	meta, err := s.BuildQuery(
		&tables,
		"tables",
		[]string{},
		columns,
		searchColumns,
		queryParams,
		[]string{},
	)

	if err != nil {
		return nil, nil, err
	}

	if tables == nil {
		tables = []types.Table{}
	}

	return tables, meta, nil
}

func (s *service) DeleteTable(id string) error {
	query, args, err := QB.Delete("tables").Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}

func (s *service) GetTableByID(id string) (types.Table, error) {
	var table types.Table
	query, args, err := QB.Select("*").From("tables").Where("id = ?", id).ToSql()
	if err != nil {
		return table, err
	}
	err = s.db.Get(&table, query, args...)
	return table, err
}

func (s *service) ParseTableFromForm(r *http.Request) (types.Table, error) {
	var table types.Table
	var err error

	table.Name = r.FormValue("name")
	if table.Name == "" {
		return table, helpers.NewValidationError("Name is required")
	}

	vendorID := r.FormValue("vendor_id")
	if vendorID == "" {
		return table, helpers.NewValidationError("Vendor ID is required")
	}
	if table.VendorId, err = uuid.Parse(vendorID); err != nil {
		return table, helpers.NewValidationError("Invalid vendor id")
	}

	if customerID := r.FormValue("customer_id"); customerID != "" {
		if table.CustomerId, err = uuid.Parse(customerID); err != nil {
			return table, helpers.NewValidationError("Invalid customer id")
		}
	}

	table.IsAvailable = helpers.ParseBoolWithDefault(r.FormValue("is_available"), true)
	table.NeedsService = helpers.ParseBoolWithDefault(r.FormValue("is_needs_service"), false)

	table.ID = uuid.New()
	return table, nil
}

func (s *service) InsertTable(table *types.Table) error {
	query, args, err := QB.Insert("tables").
		Columns("id", "name", "vendor_id", "customer_id", "is_available", "is_needs_service").
		Values(table.ID, table.Name, table.VendorId, table.CustomerId, table.IsAvailable, table.NeedsService).
		ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}

func (s *service) UpdateTableFromForm(table *types.Table, r *http.Request) error {
	if name := r.FormValue("name"); name != "" {
		table.Name = name
	}

	if vendorID := r.FormValue("vendor_id"); vendorID != "" {
		if id, err := uuid.Parse(vendorID); err == nil {
			table.VendorId = id
		}
	}

	if customerID := r.FormValue("customer_id"); customerID != "" {
		if id, err := uuid.Parse(customerID); err == nil {
			table.CustomerId = id
		}
	}

	if isAvailable := r.FormValue("is_available"); isAvailable != "" {
		table.IsAvailable = helpers.ParseBoolWithDefault(isAvailable, table.IsAvailable)
	}

	if isNeedsService := r.FormValue("is_needs_service"); isNeedsService != "" {
		table.NeedsService = helpers.ParseBoolWithDefault(isNeedsService, table.NeedsService)
	}

	return nil
}

func (s *service) UpdateTable(table *types.Table) error {
	query, args, err := QB.Update("tables").
		Set("name", table.Name).
		Set("vendor_id", table.VendorId).
		Set("customer_id", table.CustomerId).
		Set("is_available", table.IsAvailable).
		Set("is_needs_service", table.NeedsService).
		Where("id = ?", table.ID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, args...)
	return err
}
