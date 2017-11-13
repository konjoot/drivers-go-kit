package datastore_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
	"github.com/lib/pq"
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

func TestDriversUpsertBatch(t *testing.T) {

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

	dStore, err := store.NewDriversStore(db)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var drivers []*store.Driver
	drivers = []*store.Driver{
		{
			ID:            1,
			Name:          "First",
			LicenseNumber: "11-222-33",
		},
		{
			ID:            2,
			Name:          "Second",
			LicenseNumber: "11-222-334",
		},
		{
			ID:            3,
			Name:          "Third",
			LicenseNumber: "11-222-335",
		},
	}
	err = dStore.UpsertBatch(context.Background(), drivers)
	if err != nil {
		t.Error(err)
	}
	var row *sql.Row
	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 1`)
	driver1 := &store.Driver{}
	err = row.Scan(&driver1.ID, &driver1.Name, &driver1.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver1.ID =>", driver1.ID)
	if driver1.ID != 1 {
		t.Error("Expected =>", 1)
	}
	t.Log("driver1.Name =>", driver1.Name)
	if driver1.Name != "First" {
		t.Error("Expected =>", "First")
	}
	t.Log("driver1.LicenseNumber =>", driver1.LicenseNumber)
	if driver1.LicenseNumber != "11-222-33" {
		t.Error("Expected =>", "11-222-33")
	}

	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 2`)
	driver2 := &store.Driver{}
	err = row.Scan(&driver2.ID, &driver2.Name, &driver2.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver2.ID =>", driver2.ID)
	if driver2.ID != 2 {
		t.Error("Expected =>", 2)
	}
	t.Log("driver2.Name =>", driver2.Name)
	if driver2.Name != "Second" {
		t.Error("Expected =>", "Second")
	}
	t.Log("driver2.LicenseNumber =>", driver2.LicenseNumber)
	if driver2.LicenseNumber != "11-222-334" {
		t.Error("Expected =>", "11-222-334")
	}

	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 3`)
	driver3 := &store.Driver{}
	err = row.Scan(&driver3.ID, &driver3.Name, &driver3.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver3.ID =>", driver3.ID)
	if driver3.ID != 3 {
		t.Error("Expected =>", 3)
	}
	t.Log("driver3.Name =>", driver3.Name)
	if driver3.Name != "Third" {
		t.Error("Expected =>", "Third")
	}
	t.Log("driver3.LicenseNumber =>", driver3.LicenseNumber)
	if driver3.LicenseNumber != "11-222-335" {
		t.Error("Expected =>", "11-222-335")
	}

	// upsert scenario

	drivers = []*store.Driver{
		{
			ID:            1,
			Name:          "FirstUpdated",
			LicenseNumber: "11-222-44",
		},
		{
			ID:            2,
			Name:          "SecondUpdated",
			LicenseNumber: "11-222-444",
		},
		{
			ID:            3,
			Name:          "ThirdUpdated",
			LicenseNumber: "11-222-445",
		},
	}
	err = dStore.UpsertBatch(context.Background(), drivers)
	if err != nil {
		t.Error(err)
	}
	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 1`)
	driver1 = &store.Driver{}
	err = row.Scan(&driver1.ID, &driver1.Name, &driver1.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver1.ID =>", driver1.ID)
	if driver1.ID != 1 {
		t.Error("Expected =>", 1)
	}
	t.Log("driver1.Name =>", driver1.Name)
	if driver1.Name != "FirstUpdated" {
		t.Error("Expected =>", "FirstUpdated")
	}
	t.Log("driver1.LicenseNumber =>", driver1.LicenseNumber)
	if driver1.LicenseNumber != "11-222-44" {
		t.Error("Expected =>", "11-222-44")
	}

	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 2`)
	driver2 = &store.Driver{}
	err = row.Scan(&driver2.ID, &driver2.Name, &driver2.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver2.ID =>", driver2.ID)
	if driver2.ID != 2 {
		t.Error("Expected =>", 2)
	}
	t.Log("driver2.Name =>", driver2.Name)
	if driver2.Name != "SecondUpdated" {
		t.Error("Expected =>", "SecondUpdated")
	}
	t.Log("driver2.LicenseNumber =>", driver2.LicenseNumber)
	if driver2.LicenseNumber != "11-222-444" {
		t.Error("Expected =>", "11-222-444")
	}

	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 3`)
	driver3 = &store.Driver{}
	err = row.Scan(&driver3.ID, &driver3.Name, &driver3.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver3.ID =>", driver3.ID)
	if driver3.ID != 3 {
		t.Error("Expected =>", 3)
	}
	t.Log("driver3.Name =>", driver3.Name)
	if driver3.Name != "ThirdUpdated" {
		t.Error("Expected =>", "ThirdUpdated")
	}
	t.Log("driver3.LicenseNumber =>", driver3.LicenseNumber)
	if driver3.LicenseNumber != "11-222-445" {
		t.Error("Expected =>", "11-222-445")
	}

	// LicenseNumber uniq constraint violation scenario

	drivers = []*store.Driver{
		{
			ID:            1,
			Name:          "FirstConstraint",
			LicenseNumber: "11-222-44",
		},
		{
			ID:            2,
			Name:          "SecondConstraint",
			LicenseNumber: "11-222-44",
		},
		{
			ID:            3,
			Name:          "ThirdConstraint",
			LicenseNumber: "11-222-44",
		},
	}
	err = dStore.UpsertBatch(context.Background(), drivers)
	t.Log("err =>", err)
	if e, ok := err.(*pq.Error); ok {
		t.Log("e.Code =>", "23505")
		t.Log("e.Constraint =>", "drivers_license_number_key")
		if e.Code != "23505" || e.Constraint != "drivers_license_number_key" {
			t.Error("Expected e.Code =>", "23505")
			t.Error("Expected e.Constraint =>", "drivers_license_number_key")
		}
	} else {
		t.Error("Expected *pq.Error")
	}
	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 1`)
	driver1 = &store.Driver{}
	err = row.Scan(&driver1.ID, &driver1.Name, &driver1.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver1.ID =>", driver1.ID)
	if driver1.ID != 1 {
		t.Error("Expected =>", 1)
	}
	t.Log("driver1.Name =>", driver1.Name)
	if driver1.Name != "FirstUpdated" {
		t.Error("Expected =>", "FirstUpdated")
	}
	t.Log("driver1.LicenseNumber =>", driver1.LicenseNumber)
	if driver1.LicenseNumber != "11-222-44" {
		t.Error("Expected =>", "11-222-44")
	}

	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 2`)
	driver2 = &store.Driver{}
	err = row.Scan(&driver2.ID, &driver2.Name, &driver2.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver2.ID =>", driver2.ID)
	if driver2.ID != 2 {
		t.Error("Expected =>", 2)
	}
	t.Log("driver2.Name =>", driver2.Name)
	if driver2.Name != "SecondUpdated" {
		t.Error("Expected =>", "SecondUpdated")
	}
	t.Log("driver2.LicenseNumber =>", driver2.LicenseNumber)
	if driver2.LicenseNumber != "11-222-444" {
		t.Error("Expected =>", "11-222-444")
	}

	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 3`)
	driver3 = &store.Driver{}
	err = row.Scan(&driver3.ID, &driver3.Name, &driver3.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver3.ID =>", driver3.ID)
	if driver3.ID != 3 {
		t.Error("Expected =>", 3)
	}
	t.Log("driver3.Name =>", driver3.Name)
	if driver3.Name != "ThirdUpdated" {
		t.Error("Expected =>", "ThirdUpdated")
	}
	t.Log("driver3.LicenseNumber =>", driver3.LicenseNumber)
	if driver3.LicenseNumber != "11-222-445" {
		t.Error("Expected =>", "11-222-445")
	}

	// Sql injection scenario

	drivers = []*store.Driver{
		{
			ID:            1,
			Name:          "; drop table drivers; --",
			LicenseNumber: "; drop table drivers; --",
		},
	}
	err = dStore.UpsertBatch(context.Background(), drivers)
	if err != nil {
		t.Error(err)
	}
	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 1`)
	driver1 = &store.Driver{}
	err = row.Scan(&driver1.ID, &driver1.Name, &driver1.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver1.ID =>", driver1.ID)
	if driver1.ID != 1 {
		t.Error("Expected =>", 1)
	}
	t.Log("driver1.Name =>", driver1.Name)
	if driver1.Name != "; drop table drivers; --" {
		t.Error("Expected =>", "; drop table drivers; --")
	}
	t.Log("driver1.LicenseNumber =>", driver1.LicenseNumber)
	if driver1.LicenseNumber != "; drop table drivers; --" {
		t.Error("Expected =>", "; drop table drivers; --")
	}

	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 2`)
	driver2 = &store.Driver{}
	err = row.Scan(&driver2.ID, &driver2.Name, &driver2.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver2.ID =>", driver2.ID)
	if driver2.ID != 2 {
		t.Error("Expected =>", 2)
	}
	t.Log("driver2.Name =>", driver2.Name)
	if driver2.Name != "SecondUpdated" {
		t.Error("Expected =>", "SecondUpdated")
	}
	t.Log("driver2.LicenseNumber =>", driver2.LicenseNumber)
	if driver2.LicenseNumber != "11-222-444" {
		t.Error("Expected =>", "11-222-444")
	}

	row = db.QueryRow(`SELECT id, name, license_number FROM drivers where id = 3`)
	driver3 = &store.Driver{}
	err = row.Scan(&driver3.ID, &driver3.Name, &driver3.LicenseNumber)
	if err != nil {
		t.Error(err)
	}
	t.Log("driver3.ID =>", driver3.ID)
	if driver3.ID != 3 {
		t.Error("Expected =>", 3)
	}
	t.Log("driver3.Name =>", driver3.Name)
	if driver3.Name != "ThirdUpdated" {
		t.Error("Expected =>", "ThirdUpdated")
	}
	t.Log("driver3.LicenseNumber =>", driver3.LicenseNumber)
	if driver3.LicenseNumber != "11-222-445" {
		t.Error("Expected =>", "11-222-445")
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
