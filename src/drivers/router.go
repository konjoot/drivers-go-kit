package drivers

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func New(logger log.Logger) http.Handler {
	var (
		svc     driversService
		router  = mux.NewRouter()
		options = []httptransport.ServerOption{
			httptransport.ServerErrorLogger(logger),
			httptransport.ServerErrorEncoder(encodeError),
		}
	)

	router.Methods("POST").Path("/drivers/").Handler(httptransport.NewServer(
		makeDriversImportEndpoint(svc),
		decodeDriversImportRequest,
		encodeResponse,
	))
	router.Methods("GET").Path("/drivers/{id}").Handler(httptransport.NewServer(
		makeDriversGetByIDEndpoint(svc),
		decodeDriversGetByIDRequest,
		encodeResponse,
	))

	return router
}
