package datastore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// DriversStore is an interface for datastore of Drivers
type DriversStore interface {
	UpsertBatch(context.Context, []*Driver) error
	GetByID(context.Context, uint64) (*Driver, error)
}

// Driver is a struct for driver representation
type Driver struct {
	ID            uint64 `json:"id"`
	Name          string `json:"name"`
	LicenseNumber string `json:"license_number"`
}

// NewDriversStore is a constructor for DriversStore
func NewDriversStore(db *sql.DB) (DriversStore, error) {
	if db == nil {
		return nil, errors.New("*sql.DB is required")
	}
	return &driversStore{db: db}, nil
}

// driversStore is an implementation of DriversStore
type driversStore struct {
	db *sql.DB
}

// UpsertBatch prepares sql-statement with batch of drivers and applies it,
// does upsert for conflicting ids
func (ds *driversStore) UpsertBatch(ctx context.Context, drivers []*Driver) error {

	var (
		values []string
		attrs  []interface{}
	)
	for i, driver := range drivers {
		values = append(values, fmt.Sprintf("$%d, $%d, $%d", 3*i+1, 3*i+2, 3*i+3))
		attrs = append(attrs, driver.ID, driver.Name, driver.LicenseNumber)
	}
	_, err := ds.db.ExecContext(ctx,
		`INSERT INTO drivers (id, name, license_number)
		      VALUES (`+strings.Join(values, "),(")+`)
		 ON CONFLICT (id) DO UPDATE
		         SET name = EXCLUDED.name,
		             license_number = EXCLUDED.license_number`,
		attrs...,
	)
	return err
}

// GetByID selects a driver from datastore by id
func (ds *driversStore) GetByID(ctx context.Context, id uint64) (*Driver, error) {
	driver := &Driver{ID: id}
	err := ds.db.QueryRowContext(ctx,
		"SELECT name, license_number FROM drivers WHERE id = $1 LIMIT 1", id,
	).Scan(
		&driver.Name,
		&driver.LicenseNumber,
	)
	return driver, err
}
