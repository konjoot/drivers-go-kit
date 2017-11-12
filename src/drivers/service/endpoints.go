package service

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func MakeDriversImportEndpoint(svc DriversService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(driversImportRequest)
		err := svc.Import(ctx, req.Drivers)
		return emptyResponse{}, err
	}
}

func MakeDriversGetByIDEndpoint(svc DriversService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(driversGetByIDRequest)
		return svc.GetByID(ctx, req.ID)
	}
}
