package cache

import (
	_ "bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewResponseRecorder(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := NewResponseRecorder(w)
	if recorder.ResponseWriter != w {
		t.Errorf("ResponseWriter not set correctly")
	}
}

func TestWrite(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := NewResponseRecorder(w)
	data := []byte("test data")
	n, err := recorder.Write(data)
	if err != nil || n != len(data) {
		t.Errorf("Error writing data")
	}
}

func TestCopyHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := NewResponseRecorder(w)
	w.Header().Set("Test-Header", "Test-Value")
	recorder.copyHeaders()
	if recorder.headers.Get("Test-Header") != "Test-Value" {
		t.Errorf("Headers not copied correctly")
	}
}

func TestWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := NewResponseRecorder(w)
	recorder.WriteHeader(http.StatusOK)
	if recorder.status != http.StatusOK {
		t.Errorf("Status code not set correctly")
	}
}

func TestResult(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := NewResponseRecorder(w)
	recorder.WriteHeader(http.StatusOK)
	_, err := recorder.Write([]byte("test data"))
	if err != nil {
		t.Errorf("Error on write: %v", err)
	}
	result := recorder.Result()
	if result.StatusCode != http.StatusOK || string(result.Body) != "test data" {
		t.Errorf("Incorrect CacheEntry returned")
	}
}

