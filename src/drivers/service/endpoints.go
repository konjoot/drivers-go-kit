package service

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// MakeDriversImportEndpoint connects router handler with
// Import method of DriversService
func MakeDriversImportEndpoint(svc DriversService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(driversImportRequest)
		err := svc.Import(ctx, req.Drivers)
		return emptyResponse{}, err
	}
}

// MakeDriversGetByIDEndpoint connects router handler with
// GetByID method of DriversService
func MakeDriversGetByIDEndpoint(svc DriversService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(driversGetByIDRequest)
		return svc.GetByID(ctx, req.ID)
	}
}
