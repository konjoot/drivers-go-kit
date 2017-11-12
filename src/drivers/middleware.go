package drivers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

const RequestIDName = "X-Request-ID"

func logRecoverMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (out interface{}, err error) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Log("request_id", ctx.Value(RequestIDName), "panic", rec)
					err = fmt.Errorf("%s", rec)
				}
			}()
			out, err = next(ctx, request)
			if err != nil {
				logger.Log("request_id", ctx.Value(RequestIDName), "err", err)
			}

			return out, err
		}
	}
}

type requestIDMiddleware struct {
	srv http.Handler
}

func (rm *requestIDMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(
		r.Context(),
		RequestIDName,
		r.Header.Get(RequestIDName),
	))

	rm.srv.ServeHTTP(w, r)
}
