package datastore_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

func TestDriversGetByID(t *testing.T) {
	dbName, db, err := prepareTestDB()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	defer func() {
		if err := dropTestDB(dbName); err != nil {
			t.Error(err)
		}
	}()
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO drivers (id, name, license_number)
		     VALUES (1, 'First', '11-222-33'),
		            (2, 'Second', '11-222-34'),
		            (3, 'Third', '11-222-35')`,
	)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	dStore, err := store.NewDriversStore(db)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var driver *store.Driver
	driver, err = dStore.GetByID(context.Background(), 1)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver.ID =>", driver.ID)
	if driver.ID != 1 {
		t.Error("Expected =>", 1)
	}
	t.Log("driver.Name =>", driver.Name)
	if driver.Name != "First" {
		t.Error("Expected =>", "First")
	}
	t.Log("driver.LicenseNumber =>", driver.LicenseNumber)
	if driver.LicenseNumber != "11-222-33" {
		t.Error("Expected =>", "11-222-33")
	}

	driver, err = dStore.GetByID(context.Background(), 3)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver.ID =>", driver.ID)
	if driver.ID != 3 {
		t.Error("Expected =>", 3)
	}
	t.Log("driver.Name =>", driver.Name)
	if driver.Name != "Third" {
		t.Error("Expected =>", "Third")
	}
	t.Log("driver.LicenseNumber =>", driver.LicenseNumber)
	if driver.LicenseNumber != "11-222-35" {
		t.Error("Expected =>", "11-222-35")
	}

	driver, err = dStore.GetByID(context.Background(), 0)
	t.Log("err =>", err)
	if err != sql.ErrNoRows {
		t.Error("Expected =>", sql.ErrNoRows)
	}
	t.Log("driver.ID =>", driver.ID)
	if driver.ID != 0 {
		t.Error("Expected =>", 0)
	}
	t.Log("driver.Name =>", driver.Name)
	if driver.Name != "" {
		t.Error("Expected =>", "")
	}
	t.Log("driver.LicenseNumber =>", driver.LicenseNumber)
	if driver.LicenseNumber != "" {
		t.Error("Expected =>", "")
	}
}

func prepareTestDB() (string, *sql.DB, error) {
	dbName := "drivers_test_" + uuid.New().String()
	db, err := sql.Open("postgres", "postgres://drivers@localhost?sslmode=disable")
	if err != nil {
		return dbName, nil, err
	}
	defer db.Close()

	_, err = db.Exec(`CREATE DATABASE "` + dbName + `" OWNER = drivers`)
	if err != nil {
		return dbName, nil, err
	}

	db.Close()

	db, err = sql.Open("postgres", "postgres://drivers@localhost/"+dbName+"?sslmode=disable")
	if err != nil {
		return dbName, nil, err
	}

	migrate.SetTable("migrations")
	migrations := &migrate.FileMigrationSource{
		Dir: "./../migrations",
	}

	_, err = migrate.ExecMax(db, "postgres", migrations, migrate.Up, 0)
	if err != nil {
		return dbName, nil, err
	}

	return dbName, db, nil
}

func dropTestDB(dbName string) error {

	db, err := sql.Open("postgres", "postgres://drivers@localhost?sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`DROP DATABASE "` + dbName + `"`)
	return err
}
