package drivers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
	"github.com/konjoot/drivers-go-kit/src/drivers/service"
)

var (
	ErrHandlerNotFound  = errors.New("handler for the route is not found")
	ErrMethodNotAllowed = errors.New("method is not allowed")
)

func New(logger log.Logger, db store.DriversStore) http.Handler {
	var (
		svc     = service.NewDriversService(db)
		options = []httptransport.ServerOption{
			httptransport.ServerErrorLogger(logger),
			httptransport.ServerErrorEncoder(encodeError),
		}
	)

	router := mux.NewRouter()
	router.Methods("POST").Path("/import").Handler(httptransport.NewServer(
		service.MakeDriversImportEndpoint(svc),
		service.DecodeDriversImportRequest,
		encodeResponse,
		options...,
	))
	router.Methods("GET").Path("/driver/{id}").Handler(httptransport.NewServer(
		service.MakeDriversGetByIDEndpoint(svc),
		service.DecodeDriversGetByIDRequest,
		encodeResponse,
		options...,
	))
	router.NotFoundHandler = notFoundHandler{}
	router.MethodNotAllowedHandler = methodNotAllowedHandler{}

	return router
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrHandlerNotFound:
		return http.StatusNotFound
	case ErrMethodNotAllowed:
		return http.StatusMethodNotAllowed
	}

	if serr, ok := err.(statuser); ok {
		return serr.Status()
	}

	return http.StatusInternalServerError
}

type statuser interface {
	Status() int
}

type notFoundHandler struct{}

func (notFoundHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	encodeError(context.Background(), ErrHandlerNotFound, w)
}

type methodNotAllowedHandler struct{}

func (methodNotAllowedHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	encodeError(context.Background(), ErrMethodNotAllowed, w)
}
