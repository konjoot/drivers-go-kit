package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
)

var (
	ErrDriverNotFound = errors.New("driver is not found")
	ErrEmptySet       = errors.New("empty set")
	ErrZeroID         = errors.New("invalid id; should be greater then 0")
)

var regexpString = `^[0-9]{2}-[0-9]{3}-[0-9]{2}$`
var validLicenseNumber = regexp.MustCompile(regexpString)

// DriversService is an interface for "Drivers" service
type DriversService interface {
	Import(context.Context, []*store.Driver) error
	GetByID(context.Context, uint64) (*store.Driver, error)
}

func NewDriversService(db store.DriversStore) DriversService {
	return &driversService{store: db}
}

type driversService struct {
	store store.DriversStore
}

func (drs *driversService) Import(ctx context.Context, drivers []*store.Driver) error {
	if len(drivers) < 1 {
		return &statusError{http.StatusBadRequest, ErrEmptySet}
	}

	for _, driver := range drivers {
		if driver.ID == 0 {
			return &statusError{http.StatusBadRequest, ErrZeroID}
		}

		nameRunesLen := len([]rune(driver.Name))
		if nameRunesLen < 4 || nameRunesLen > 1000 {
			return &statusError{
				http.StatusBadRequest,
				fmt.Errorf(
					"invalid name length; should be from 4 to 1000 UTF symbols, but not %d",
					nameRunesLen,
				),
			}
		}

		if !validLicenseNumber.MatchString(driver.LicenseNumber) {
			return &statusError{
				http.StatusBadRequest,
				fmt.Errorf(
					"invalid license_number format; should match %s, but was %s",
					regexpString,
					driver.LicenseNumber,
				),
			}
		}
	}

	return drs.store.UpsertBatch(ctx, drivers)
}

func (drs *driversService) GetByID(ctx context.Context, id uint64) (*store.Driver, error) {
	if id == 0 {
		return nil, &statusError{http.StatusBadRequest, ErrZeroID}
	}

	driver, err := drs.store.GetByID(ctx, id)

	if err == sql.ErrNoRows {
		return nil, &statusError{http.StatusNotFound, err}
	}

	return driver, err
}

type statusError struct {
	status int
	err    error
}

func (se *statusError) Error() string {
	if se.err == nil {
		return ""
	}
	return se.err.Error()
}

func (se *statusError) Status() int {
	return se.status
}
