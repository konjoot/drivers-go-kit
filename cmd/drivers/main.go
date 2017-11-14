// Runner of the Drivers application.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/go-kit/kit/log"
	"github.com/konjoot/drivers-go-kit/src/drivers"
	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
	migrate "github.com/rubenv/sql-migrate"
)

// Drivers app constructs, runs and stops here
func main() {

	// prepare default values from ENV

	// DATABASE_URL used on Heroku to connect to PostgreSQL
	defaultDbURL := os.Getenv("DATABASE_URL")
	if defaultDbURL == "" {
		defaultDbURL = "postgres://drivers@localhost/drivers_dev?sslmode=disable"
	}

	// PORT used on Heroku for HTTP-listening
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "8080"
	}

	// application flags
	httpAddr := flag.String("http.addr",
		":"+defaultPort,
		"HTTP listen address",
	)
	dbURL := flag.String("db.url",
		defaultDbURL,
		"DB connection URL",
	)
	dbPoolSize := flag.Int("db.pool_size",
		16,
		"Number of idle connections allowed",
	)
	flag.Parse()

	// Logger initialization
	logger := log.NewLogfmtLogger(os.Stderr)

	// DB connection initialization
	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		logger.Log("func", "sql.Open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(*dbPoolSize * 3 / 2)
	db.SetMaxIdleConns(*dbPoolSize)

	if err = db.Ping(); err != nil {
		logger.Log("func", "db.Ping", "err", err)
		os.Exit(1)
	}

	// apply migrations
	migrate.SetTable("migrations")
	migrations := &migrate.FileMigrationSource{
		Dir: "./src/drivers/migrations",
	}

	n, err := migrate.ExecMax(db, "postgres", migrations, migrate.Up, 0)
	if err != nil {
		logger.Log("func", "migrate.ExecMax", "err", err)
		os.Exit(1)
	}
	if n == 1 {
		logger.Log("message", fmt.Sprintf("%d migration applied", n))
	} else {
		logger.Log("message", fmt.Sprintf("%d migrations applied", n))
	}

	// DriversStore initialization
	dStore, err := store.NewDriversStore(db)
	if err != nil {
		logger.Log("func", "store.NewDriversStore", "err", err)
		os.Exit(1)
	}

	// HTTP-handler initialization
	handler := &server{
		assets: http.FileServer(http.Dir("./build")),
		api:    drivers.New(logger, dStore),
	}

	// HTTP-server initialization
	srv := &http.Server{
		Addr:    *httpAddr,
		Handler: handler,
	}

	// run the world
	errs := make(chan error, 1)
	go func() {
		logger.Log("message", "HTTP-server is listening on "+(*httpAddr))
		errs <- srv.ListenAndServe()
	}()

	// listen os.Interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// stop the app if
	var exitCode int
	select {
	case <-stop:
		// os.Interrupt signal appeared
	case err = <-errs:
		// HTTP-server accedentally stops
		logger.Log("func", "srv.ListenAndServe", "err", err)
		exitCode = 1
	}

	// shutting app down
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		logger.Log("func", "srv.Shutdown", "err", err)
	}

	logger.Log("message", "service is gracefully stopped")

	os.Exit(exitCode)
}

// server is a decorator for http.Handler
// it provides capabilities to serve API and
// static files for API documentation
type server struct {
	assets http.Handler
	api    http.Handler
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.api.ServeHTTP(w, r)
		return
	}

	s.assets.ServeHTTP(w, r)
}
