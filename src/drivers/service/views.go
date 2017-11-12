package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
)

type driversGetByIDRequest struct {
	ID uint64 `json:"id"`
}

func DecodeDriversGetByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	idString, ok := vars["id"]
	if !ok {
		return nil, errors.New("Bad routing")
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, err
	}
	return driversGetByIDRequest{ID: uint64(id)}, nil
}

type driversImportRequest struct {
	Drivers []*store.Driver
}

func DecodeDriversImportRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request driversImportRequest
	if err := json.NewDecoder(r.Body).Decode(&request.Drivers); err != nil {
		return nil, err
	}
	return request, nil
}

type emptyResponse struct{}
