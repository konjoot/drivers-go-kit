package drivers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// ctxKey type is needed to avoid
// key collisions in context
type ctxKey string

const (
	requestIDName ctxKey = "X-Request-ID"
)

// logRecoverMiddleware wraps endpoints to provide logging of errors
// and panic recovery
func logRecoverMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (out interface{}, err error) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Log("request_id", ctx.Value(requestIDName), "panic", rec)
					err = fmt.Errorf("%s", rec)
				}
			}()
			out, err = next(ctx, request)
			if err != nil {
				logger.Log("request_id", ctx.Value(requestIDName), "err", err)
			}

			return out, err
		}
	}
}

// requestIDMiddleware decorates http.Handler
// get X-Request-ID (provided by Heroku)
// from request Headers and
// stores it into request's context
type requestIDMiddleware struct {
	srv http.Handler
}

func (rm *requestIDMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(
		r.Context(),
		requestIDName,
		r.Header.Get(string(requestIDName)),
	))

	rm.srv.ServeHTTP(w, r)
}
