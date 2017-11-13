package drivers_test

import (
	"bytes"
	"context"
	"database/sql"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/konjoot/drivers-go-kit/src/drivers"
	store "github.com/konjoot/drivers-go-kit/src/drivers/datastore"
)

func TestDrivers(t *testing.T) {
	inMemStore := &inMemStorage{
		db: make(map[uint64]*store.Driver),
	}
	srv := drivers.New(nopLogger{}, inMemStore)

	request := httptest.NewRequest("POST",
		"/api/import",
		bytes.NewBuffer([]byte(`[{"id":1,"name":"John","license_number":"11-222-33"}]`)),
	)
	response := httptest.NewRecorder()

	srv.ServeHTTP(response, request)

	contentType := response.Header().Get("Content-Type")
	t.Log("response Content-Type =>", contentType)
	if contentType != "application/json; charset=utf-8" {
		t.Error("Expected =>", "application/json; charset=utf-8")
	}

	t.Log("response status =>", response.Code)
	if response.Code != http.StatusOK {
		t.Error("Expected =>", http.StatusOK)
	}

	bts, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error(err)
	}
	t.Log("response body =>", string(bts))
	if string(bts) != "{}\n" {
		t.Error("Expected =>", "{}\n")
	}

	request = httptest.NewRequest("GET", "/api/driver/1", nil)
	response = httptest.NewRecorder()

	srv.ServeHTTP(response, request)

	contentType = response.Header().Get("Content-Type")
	t.Log("response Content-Type =>", contentType)
	if contentType != "application/json; charset=utf-8" {
		t.Error("Expected =>", "application/json; charset=utf-8")
	}

	t.Log("response status =>", response.Code)
	if response.Code != http.StatusOK {
		t.Error("Expected =>", http.StatusOK)
	}

	bts, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error(err)
	}
	t.Log("response body =>", string(bts))
	if string(bts) != `{"id":1,"name":"John","license_number":"11-222-33"}`+"\n" {
		t.Error("Expected =>", `{"id":1,"name":"John","license_number":"11-222-33"}`+"\n")
	}

	request = httptest.NewRequest("GET", "/api/driver/1345", nil)
	response = httptest.NewRecorder()

	srv.ServeHTTP(response, request)

	contentType = response.Header().Get("Content-Type")
	t.Log("response Content-Type =>", contentType)
	if contentType != "application/json; charset=utf-8" {
		t.Error("Expected =>", "application/json; charset=utf-8")
	}

	t.Log("response status =>", response.Code)
	if response.Code != http.StatusNotFound {
		t.Error("Expected =>", http.StatusNotFound)
	}

	bts, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error(err)
	}
	t.Log("response body =>", string(bts))
	if string(bts) != `{"error":"status=404, error=driver with id=1345 is not found"}`+"\n" {
		t.Error("Expected =>", `{"error":"status=404, error=driver with id=1345 is not found"}`+"\n")
	}
}

type inMemStorage struct {
	store.DriversStore
	sync.RWMutex

	db map[uint64]*store.Driver
}

func (ms *inMemStorage) UpsertBatch(_ context.Context, drivers []*store.Driver) error {
	ms.Lock()
	for _, driver := range drivers {
		ms.db[driver.ID] = driver
	}
	ms.Unlock()
	return nil
}

func (ms *inMemStorage) GetByID(_ context.Context, id uint64) (*store.Driver, error) {
	ms.RLock()
	defer ms.RUnlock()

	if driver, ok := ms.db[id]; ok {
		return driver, nil
	}

	return nil, sql.ErrNoRows
}

type nopLogger struct{}

func (nopLogger) Log(...interface{}) error {
	return nil
}
