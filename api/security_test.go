package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// pixelizeBody builds a multipart body for /pixelize, the picture being sent
// as the "file" part.
func pixelizeBody(t *testing.T, slices, edge, flt, filename string, picture []byte) (*bytes.Buffer, string) {
	t.Helper()
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	for name, value := range map[string]string{"slices": slices, "edge": edge, "filter": flt} {
		if err := w.WriteField(name, value); err != nil {
			t.Fatalf("error writing field %q: %v", name, err)
		}
	}
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("error creating file part: %v", err)
	}
	if _, err := fw.Write(picture); err != nil {
		t.Fatalf("error writing the picture: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("error closing the multipart writer: %v", err)
	}
	return buf, w.FormDataContentType()
}

func postPixelize(t *testing.T, r http.Handler, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/pixelize", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)
	return w
}

func testPicture(t *testing.T) []byte {
	t.Helper()
	f, err := os.Open("img_test.png")
	if err != nil {
		t.Fatalf("cannot open img_test.png: %v", err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("cannot read img_test.png: %v", err)
	}
	return b
}

// pngHeader forges the signature and the IHDR chunk of a PNG declaring the
// given dimensions - all image.DecodeConfig needs to read.
func pngHeader(width, height uint32) []byte {
	ihdr := make([]byte, 0, 17)
	ihdr = append(ihdr, []byte("IHDR")...)
	ihdr = binary.BigEndian.AppendUint32(ihdr, width)
	ihdr = binary.BigEndian.AppendUint32(ihdr, height)
	ihdr = append(ihdr, 8, 6, 0, 0, 0) // 8 bits, truecolor + alpha

	b := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}
	b = binary.BigEndian.AppendUint32(b, 13) // IHDR length
	b = append(b, ihdr...)
	return binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(ihdr))
}

func TestPixelizeRejectsHugeUpload(t *testing.T) {
	defer func(v int64) { maxUploadBytes = v }(maxUploadBytes)
	maxUploadBytes = 64 << 10 // 64 KiB
	r := setupRouter()

	body, ct := pixelizeBody(t, "100", "short", "cga4", "big.png", bytes.Repeat([]byte{0x42}, int(maxUploadBytes)+1))
	w := postPixelize(t, r, body, ct)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("incorrect status code, got %d, want %d, body: %s", w.Code, http.StatusRequestEntityTooLarge, w.Body)
	}
}

func TestPixelizeRejectsDecompressionBomb(t *testing.T) {
	r := setupRouter()

	// 30000 x 30000 = 900 MP, 3.6 GB once decoded, from a 33 bytes header
	body, ct := pixelizeBody(t, "100", "short", "cga4", "bomb.png", pngHeader(30000, 30000))
	w := postPixelize(t, r, body, ct)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("incorrect status code, got %d, want %d, body: %s", w.Code, http.StatusRequestEntityTooLarge, w.Body)
	}
	if !strings.Contains(w.Body.String(), "too large") {
		t.Errorf("incorrect body, got %q, want it to mention the size limit", w.Body.String())
	}
}

func TestPixelizeRejectsUnsupportedFormat(t *testing.T) {
	r := setupRouter()

	body, ct := pixelizeBody(t, "100", "short", "cga4", "not.png", []byte("definitely not a picture"))
	w := postPixelize(t, r, body, ct)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("incorrect status code, got %d, want %d, body: %s", w.Code, http.StatusUnsupportedMediaType, w.Body)
	}
}

func TestPixelizeRejectsTooManySlices(t *testing.T) {
	r := setupRouter()

	// img_test.png is 240 x 320
	body, ct := pixelizeBody(t, "1000", "short", "cga4", "img_test.png", testPicture(t))
	w := postPixelize(t, r, body, ct)

	if w.Code != http.StatusBadRequest {
		t.Errorf("incorrect status code, got %d, want %d, body: %s", w.Code, http.StatusBadRequest, w.Body)
	}
}

func TestPixelizeErrorsDoNotLeakInternals(t *testing.T) {
	r := setupRouter()

	body, ct := pixelizeBody(t, "100", "short", "cga4", "not.png", []byte("definitely not a picture"))
	w := postPixelize(t, r, body, ct)

	for _, leak := range []string{"go-pixelart", "goroutine", ".go:"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Errorf("response leaks internals (%q): %s", leak, w.Body.String())
		}
	}
}

func TestCORSHeaders(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/pixelize", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("incorrect status code, got %d, want %d", w.Code, http.StatusNoContent)
	}
	h := w.Result().Header
	if h.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("incorrect Access-Control-Allow-Origin, got %q, want %q", h.Get("Access-Control-Allow-Origin"), "*")
	}
	// "*" together with credentials is rejected by browsers and would open
	// credentialed cross origin calls
	if v := h.Get("Access-Control-Allow-Credentials"); v != "" {
		t.Errorf("Access-Control-Allow-Credentials must not be set with a wildcard origin, got %q", v)
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := setupRouter()

	for _, path := range []string{"/", "/health", "/favicon.ico"} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

			h := w.Result().Header
			for name, want := range map[string]string{
				"X-Content-Type-Options": "nosniff",
				"X-Frame-Options":        "DENY",
				"Referrer-Policy":        "no-referrer",
			} {
				if h.Get(name) != want {
					t.Errorf("incorrect %s, got %q, want %q", name, h.Get(name), want)
				}
			}
			if !strings.Contains(h.Get("Content-Security-Policy"), "default-src 'self'") {
				t.Errorf("incorrect Content-Security-Policy, got %q", h.Get("Content-Security-Policy"))
			}
		})
	}
}

func TestEnvInt64(t *testing.T) {
	tt := []struct {
		name, value string
		def, want   int64
	}{
		{"unset", "", 42, 42},
		{"valid", "1024", 42, 1024},
		{"not a number", "many", 42, 42},
		{"negative", "-1", 42, 42},
		{"zero", "0", 42, 42},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PIXELART_TEST_LIMIT", tc.value)
			if got := envInt64("PIXELART_TEST_LIMIT", tc.def); got != tc.want {
				t.Errorf("envInt64(%q, %d) = %d, want %d", tc.value, tc.def, got, tc.want)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	if got, want := sanitize("ok\nfake log line\r\x00"), "okfake log line"; got != want {
		t.Errorf("sanitize() = %q, want %q", got, want)
	}
}
