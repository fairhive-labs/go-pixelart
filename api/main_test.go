package main

import (
	"bytes"
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
		t.Errorf("cannot open file %q: %v\n", path, err)
		t.FailNow()
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		t.Errorf("cannot copy file %q: %v\n", path, err)
		t.FailNow()
	}

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pixelize?mime=json", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf(w.Body.String())
		t.Errorf("incorrect status code, got %d, want %d\n", w.Code, http.StatusOK)
		t.FailNow()
	}
}
