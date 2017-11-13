package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
	"github.com/lib/pq"
)

var (
	ErrDriverNotFound = errors.New("driver is not found")
	ErrEmptySet       = errors.New("empty set")
	ErrZeroID         = errors.New("invalid id; should be greater then 0")
)

var (
	ErrInvalidLengthTempl           = "invalid length; field %s should be from %d to %d UTF-8 symbols, but not %d"
	ErrInvalidFormatTempl           = "invalid format; %s field should match %s, but was %s"
	ErrNotFoundTempl                = "%s with %s=%d is not found"
	ErrInvalidCollectionLengthTempl = "invalid collection length; collection %s should be from %d to %d elements, but not %d"
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

	driversLength := len(drivers)
	if driversLength < 1 || driversLength > 1000 {
		return BadRequest(fmt.Errorf(ErrInvalidCollectionLengthTempl,
			"drivers", 1, 1000, driversLength),
		)
	}

	for _, driver := range drivers {
		if err := validateDriver(driver); err != nil {
			return BadRequest(err)
		}
	}

	err := drs.store.UpsertBatch(ctx, drivers)
	if e, ok := err.(*pq.Error); ok && e.Constraint == "drivers_license_number_key" {
		return Conflict(fmt.Errorf(e.Detail))
	}
	if err != nil {
		return InternalServerError(err)
	}

	return nil
}

func (drs *driversService) GetByID(ctx context.Context, id uint64) (*store.Driver, error) {
	if id == 0 {
		return nil, BadRequest(ErrZeroID)
	}

	driver, err := drs.store.GetByID(ctx, id)

	if err == sql.ErrNoRows {
		return nil, NotFound(fmt.Errorf(ErrNotFoundTempl, "driver", "id", id))
	}

	if err != nil {
		return nil, InternalServerError(err)
	}

	return driver, nil
}

func validateDriver(driver *store.Driver) error {
	if driver.ID == 0 {
		return ErrZeroID
	}

	nameRunesLen := len([]rune(driver.Name))
	if nameRunesLen < 4 || nameRunesLen > 1000 {
		return fmt.Errorf(ErrInvalidLengthTempl,
			"name", 4, 1000, nameRunesLen)
	}

	if !validLicenseNumber.MatchString(driver.LicenseNumber) {
		return fmt.Errorf(ErrInvalidFormatTempl,
			"license_number",
			regexpString,
			driver.LicenseNumber,
		)
	}
	return nil
}

func StatusError(status int, err error) *statusError {
	return &statusError{status, err}
}

func BadRequest(err error) *statusError {
	return &statusError{http.StatusBadRequest, err}
}

func NotFound(err error) *statusError {
	return &statusError{http.StatusNotFound, err}
}

func Conflict(err error) *statusError {
	return &statusError{http.StatusConflict, err}
}

func InternalServerError(err error) *statusError {
	return &statusError{http.StatusInternalServerError, err}
}

type statusError struct {
	status int
	err    error
}

func (se *statusError) Error() string {
	if se.err == nil {
		return ""
	}
	return fmt.Sprintf("status=%d, error=%s", se.status, se.err)
}

func (se *statusError) Status() int {
	return se.status
}
