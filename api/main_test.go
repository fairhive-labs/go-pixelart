package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetIndex(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("incorrect status code, got %d, want %d\n", w.Code, http.StatusOK)
		t.FailNow()
	}
	if w.Body.Len() == 0 {
		t.Error("Body cannot be empty")
		t.FailNow()
	}

	headers := w.Result().Header
	if headers.Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("Access-Control-Allow-Methods header cannot be empty, must be %q\n", "POST, OPTIONS")
		t.FailNow()
	}
	if headers.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("incorrect Content-Type, got %q, want %q\n", headers.Get("Content-Type"), "text/html; charset=utf-8")
		t.FailNow()
	}
}

func TestGetFilter(t *testing.T) {
	if _, err := getFilter("cga2"); err != nil {
		t.Errorf("incorrect error getting filter cga2, got %v, want nil\n", err)
		t.FailNow()
	}
	if _, err := getFilter("fooFilter"); !errors.Is(err, errUnsupportedFilter) {
		t.Errorf("incorrect error getting filter fooFilter, got nil, want %v\n", errUnsupportedFilter)
		t.FailNow()
	}
}

func TestPixelize(t *testing.T) {
	r := setupRouter()

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	fw, err := writer.CreateFormField("edge")
	if err != nil {
		t.Errorf("error creating form field %q: %v", "edge", err)
		t.FailNow()
	}
	_, err = io.Copy(fw, strings.NewReader("short"))
	if err != nil {
		t.Errorf("error copying value %q in io.Writer: %v", "short", err)
		t.FailNow()
	}

	fw, err = writer.CreateFormField("slices")
	if err != nil {
		t.Errorf("error creating form field %q: %v", "slices", err)
		t.FailNow()
	}
	_, err = io.Copy(fw, strings.NewReader("100"))
	if err != nil {
		t.Errorf("error copying value %q in io.Writer: %v", "100", err)
		t.FailNow()
	}

	fw, err = writer.CreateFormField("filter")
	if err != nil {
		t.Errorf("error creating form field %q: %v", "filter", err)
		t.FailNow()
	}
	_, err = io.Copy(fw, strings.NewReader("cga4"))
	if err != nil {
		t.Errorf("error copying value %q in io.Writer: %v", "cga4", err)
		t.FailNow()
	}

	path := "img_test.png"
	fw, err = writer.CreateFormFile("file", path)
	if err != nil {
	}
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("error opening file %q: %v\n", path, err)
		t.FailNow()
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		t.Errorf("error copying file %q: %v\n", path, err)
		t.FailNow()
	}

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pixelize?mime=json", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("incorrect status code, got %d, want %d\n", w.Code, http.StatusOK)
		t.Errorf("body content: %v\n", w.Body.String())
		t.FailNow()
	}

	var res struct {
		Data     string
		Encoding string
		Filter   string
		Length   int
	}
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Errorf("error decoding body %v: %v", w.Body.String(), err)
		t.FailNow()
	}

	if res.Encoding != "base64" {
		t.Errorf("incorrect encoding, got %v, want %v\n", res.Encoding, "base64")
		t.FailNow()
	}
	if res.Filter != "cga4" {
		t.Errorf("incorrect filter, got %v, want %v\n", res.Encoding, "cga4")
		t.FailNow()
	}
	if res.Length != 4628 {
		t.Errorf("incorrect length, got %d, want %d\n", res.Length, 4628)
		t.FailNow()
	}
	if res.Data == "" {
		t.Error("data cannot be empty")
		t.FailNow()
	}
}

func TestHealth(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("incorrect status code, got %d, want %d\n", w.Code, http.StatusOK)
		t.FailNow()
	}

	headers := w.Result().Header
	if headers.Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("Access-Control-Allow-Methods header cannot be empty, must be %q\n", "POST, OPTIONS")
		t.FailNow()
	}
	if headers.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("incorrect Content-Type, got %q, want %q\n", headers.Get("Content-Type"), "text/plain; charset=utf-8")
		t.FailNow()
	}

	if w.Body.Len() == 0 {
		t.Error("Body cannot be empty")
		t.FailNow()
	}

	if w.Body.String() != "ok" {
		t.Errorf("incorrect body: must only contain %q", "ok")
		t.FailNow()
	}
}

func TestGetFavicon(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/favicon.ico", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("incorrect status code, got %d, want %d\n", w.Code, http.StatusOK)
		t.FailNow()
	}

	headers := w.Result().Header
	if headers.Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("Access-Control-Allow-Methods header cannot be empty, must be %q\n", "POST, OPTIONS")
		t.FailNow()
	}
	if headers.Get("Content-Type") != "image/x-icon" {
		t.Errorf("incorrect Content-Type, got %q, want %q\n", headers.Get("Content-Type"), "image/x-icon")
		t.FailNow()
	}

	if w.Body.Len() == 0 {
		t.Error("Body cannot be empty")
		t.FailNow()
	}
}
