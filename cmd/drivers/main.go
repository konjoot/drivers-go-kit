package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-kit/kit/log"
)

func main() {

	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	srv := &http.Server{
		Addr:    *httpAddr,
		Handler: drivers.New(),
	}

	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	logger.Log("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	errs <- srv.Shutdown(ctx)

	logger.Log("Server gracefully stopped")

	if err := <-errs; err != nil {
		logger.Log("Exit with error", err)
		os.Exit(1)
	}

	logger.Log("Exit normally")
}
