package main

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/lib/pq"

	"github.com/go-kit/kit/log"
	"github.com/konjoot/drivers-go-kit/src/drivers"
	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
)

func main() {

	httpAddr := flag.String("http.addr",
		":8080",
		"HTTP listen address",
	)
	dbURL := flag.String("db.url",
		"postgres://drivers@localhost/drivers_dev?sslmode=disable",
		"DB connection URL",
	)
	dbPoolSize := flag.Int("db.pool_size",
		16,
		"Number of idle connections allowed",
	)
	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		logger.Log("func", "sql.Open", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(*dbPoolSize * 3 / 2)
	db.SetMaxIdleConns(*dbPoolSize)

	if err = db.Ping(); err != nil {
		logger.Log("func", "db.Ping", "error", err)
		os.Exit(1)
	}

	dStore, err := store.NewDriversStore(db)
	if err != nil {
		logger.Log("func", "store.NewDriversStore", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    *httpAddr,
		Handler: drivers.New(logger, dStore),
	}

	errs := make(chan error, 1)
	go func() {
		logger.Log("message", "HTTP-server is listening on "+(*httpAddr))
		errs <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	var exitCode int
	select {
	case <-stop:
	case err = <-errs:
		logger.Log("func", "srv.ListenAndServe", "error", err)
		exitCode = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		logger.Log("func", "srv.Shutdown", "error", err)
	}

	logger.Log("message", "service is gracefully stopped")

	os.Exit(exitCode)
}
