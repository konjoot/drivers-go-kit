package service_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
	"github.com/konjoot/drivers-go-kit/src/drivers/service"
	"github.com/lib/pq"
)

func TestDriversImport(t *testing.T) {

	var (
		err    error
		srv    service.DriversService
		dbMock *mockStore
	)
	for _, tc := range []struct {
		name      string
		drivers   []*store.Driver
		importErr error
		expErr    error
	}{
		{
			name: "Success",
			drivers: []*store.Driver{
				{
					ID:            1,
					Name:          "John",
					LicenseNumber: "11-222-33",
				},
			},
		},
		{
			name:   "EmptyCollection",
			expErr: service.BadRequest(errors.New("invalid collection length; collection drivers should be from 1 to 1000 elements, but not 0")),
		},
		{
			name:    "TooBigCollection",
			drivers: make([]*store.Driver, 1001, 1001),
			expErr:  service.BadRequest(errors.New("invalid collection length; collection drivers should be from 1 to 1000 elements, but not 1001")),
		},
		{
			name: "InvalidDriverZeroID",
			drivers: []*store.Driver{
				{
					ID:            0,
					Name:          "John",
					LicenseNumber: "11-222-33",
				},
			},
			expErr: service.BadRequest(service.ErrZeroID),
		},
		{
			name: "InvalidDriverNameIsTooShort",
			drivers: []*store.Driver{
				{
					ID:            1,
					Name:          "jo",
					LicenseNumber: "11-222-33",
				},
			},
			expErr: service.BadRequest(errors.New("invalid length; field name should be from 4 to 1000 UTF-8 symbols, but not 2")),
		},
		{
			name: "InvalidDriverNameIsTooLong",
			drivers: []*store.Driver{
				{
					ID:            1,
					Name:          strings.Repeat("a", 1001),
					LicenseNumber: "11-222-33",
				},
			},
			expErr: service.BadRequest(errors.New("invalid length; field name should be from 4 to 1000 UTF-8 symbols, but not 1001")),
		},
		{
			name: "InvalidDriverLicenseNumber",
			drivers: []*store.Driver{
				{
					ID:            1,
					Name:          "John",
					LicenseNumber: "11-222-333",
				},
			},
			expErr: service.BadRequest(errors.New("invalid format; license_number field should match ^[0-9]{2}-[0-9]{3}-[0-9]{2}$, but was 11-222-333")),
		},
		{
			name: "UniqConstraintViolation",
			drivers: []*store.Driver{
				{
					ID:            1,
					Name:          "John",
					LicenseNumber: "11-222-33",
				},
			},
			importErr: &pq.Error{
				Constraint: "drivers_license_number_key",
				Detail:     "detail",
			},
			expErr: service.Conflict(errors.New("detail")),
		},
		{
			name: "InternalServerError",
			drivers: []*store.Driver{
				{
					ID:            1,
					Name:          "John",
					LicenseNumber: "11-222-33",
				},
			},
			importErr: errors.New("internal"),
			expErr:    service.InternalServerError(errors.New("internal")),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dbMock = &mockStore{importErr: tc.importErr}
			srv = service.NewDriversService(dbMock)
			err = srv.Import(context.Background(), tc.drivers)
			t.Log("err =>", err)
			if fmt.Sprint(err) != fmt.Sprint(tc.expErr) {
				t.Error("Expected =>", tc.expErr)
			}
		})
	}
}
func TestDriversGetByID(t *testing.T) {
	var (
		driver *store.Driver
		err    error
		srv    service.DriversService
		dbMock *mockStore
	)

	type statuser interface {
		Status() int
	}

	for _, tc := range []struct {
		name        string
		id          uint64
		storeDriver *store.Driver
		storeErr    error
		expErr      error
		expDriver   *store.Driver
	}{
		{
			name:        "Success",
			id:          1,
			storeDriver: &store.Driver{},
			expDriver:   &store.Driver{},
		},
		{
			name:   "ErrZeroID",
			id:     0,
			expErr: service.BadRequest(service.ErrZeroID),
		},
		{
			name:     "ErrNotFound",
			id:       1,
			storeErr: sql.ErrNoRows,
			expErr:   service.NotFound(errors.New("driver with id=1 is not found")),
		},
		{
			name:     "ErrInternalServerError",
			id:       1,
			storeErr: errors.New("internal"),
			expErr:   service.InternalServerError(errors.New("internal")),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dbMock = &mockStore{
				getByIDDriver: tc.storeDriver,
				getByIDErr:    tc.storeErr,
			}
			srv = service.NewDriversService(dbMock)
			driver, err = srv.GetByID(context.Background(), tc.id)
			t.Log("err =>", err)
			if fmt.Sprint(err) != fmt.Sprint(tc.expErr) {
				t.Error("Expected =>", tc.expErr)
			}
			t.Log("driver =>", driver)
			if fmt.Sprint(driver) != fmt.Sprint(tc.expDriver) {
				t.Error("Expected =>", tc.expDriver)
			}
		})
	}
}

type mockStore struct {
	store.DriversStore

	getByIDDriver *store.Driver
	getByIDErr    error

	importErr error
}

func (ms *mockStore) GetByID(context.Context, uint64) (*store.Driver, error) {
	return ms.getByIDDriver, ms.getByIDErr
}

func (ms *mockStore) UpsertBatch(context.Context, []*store.Driver) error {
	return ms.importErr
}
