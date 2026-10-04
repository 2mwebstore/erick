package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kongchansila/portfolio/backend/internal/handlers"
	"github.com/kongchansila/portfolio/backend/internal/uploads"
)

type memStore struct{ puts int }

func (m *memStore) Delete(context.Context, string) error { return nil }

func (m *memStore) Put(context.Context, string, string, []byte) error {
	m.puts++
	return nil
}

const pngBytes = "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"

func uploadRequest(t *testing.T, fields map[string]string, fileName, file string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	for k, v := range fields {
		_ = form.WriteField(k, v)
	}
	if fileName != "" {
		w, err := form.CreateFormFile("file", fileName)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(file))
	}
	_ = form.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/uploads/image", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	return req
}

func uploadHandler(store uploads.Store, max int64) *handlers.UploadHandler {
	var u *uploads.Uploader
	if store != nil {
		u = uploads.NewUploader(store, "https://cdn.example.com", max)
	}
	return handlers.NewUploadHandler(u, nil, nil, quietLog(), false)
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("not JSON: %q", rec.Body.String())
	}
	return body
}

func TestUploadStoresAnImageAndAnswersItsURL(t *testing.T) {
	store := &memStore{}
	rec := httptest.NewRecorder()
	// The file field first: the order of the form's fields must not matter.
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	w, _ := form.CreateFormFile("file", "screenshot.png")
	_, _ = w.Write([]byte(pngBytes))
	_ = form.WriteField("folder", "projects")
	_ = form.Close()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/uploads/image", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())

	uploadHandler(store, 1<<20).Image(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	got := decodeBody(t, rec)
	url, _ := got["url"].(string)
	if !strings.HasPrefix(url, "https://cdn.example.com/projects/") || !strings.HasSuffix(url, ".png") {
		t.Errorf("url %q", url)
	}
	if strings.Contains(url, "screenshot") {
		t.Error("the browser's file name leaked into the stored key")
	}
	if got["ok"] != true || store.puts != 1 {
		t.Errorf("ok %v, puts %d", got["ok"], store.puts)
	}
}

func TestUploadRefusalsAnswerWithAFieldError(t *testing.T) {
	cases := map[string]struct {
		req  func(t *testing.T) *http.Request
		max  int64
		code int
		key  string
	}{
		"an HTML page named photo.jpg": {
			req: func(t *testing.T) *http.Request {
				return uploadRequest(t, map[string]string{"folder": "projects"}, "photo.jpg", "<!doctype html><script>alert(1)</script>")
			},
			max: 1 << 20, code: http.StatusBadRequest, key: "file",
		},
		"an SVG": {
			req: func(t *testing.T) *http.Request {
				return uploadRequest(t, map[string]string{"folder": "profiles"}, "logo.svg", `<svg xmlns="http://www.w3.org/2000/svg"/>`)
			},
			max: 1 << 20, code: http.StatusBadRequest, key: "file",
		},
		"a file over the limit": {
			req: func(t *testing.T) *http.Request {
				return uploadRequest(t, map[string]string{"folder": "projects"}, "big.png", pngBytes+strings.Repeat("x", 2048))
			},
			max: 1024, code: http.StatusRequestEntityTooLarge, key: "file",
		},
		"no file": {
			req: func(t *testing.T) *http.Request {
				return uploadRequest(t, map[string]string{"folder": "projects"}, "", "")
			},
			max: 1 << 20, code: http.StatusBadRequest, key: "file",
		},
		"an unknown folder": {
			req: func(t *testing.T) *http.Request {
				return uploadRequest(t, map[string]string{"folder": "../../etc"}, "x.png", pngBytes)
			},
			max: 1 << 20, code: http.StatusBadRequest, key: "folder",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			store := &memStore{}
			rec := httptest.NewRecorder()
			uploadHandler(store, tc.max).Image(rec, tc.req(t))

			if rec.Code != tc.code {
				t.Fatalf("status %d, want %d: %s", rec.Code, tc.code, rec.Body.String())
			}
			errs, _ := decodeBody(t, rec)["errors"].(map[string]any)
			if errs[tc.key] == nil {
				t.Errorf("no error on %q: %s", tc.key, rec.Body.String())
			}
			if store.puts != 0 {
				t.Error("a refused file was stored")
			}
		})
	}
}

func TestUploadsAreOffWithoutR2(t *testing.T) {
	h := uploadHandler(nil, 0)

	rec := httptest.NewRecorder()
	h.Image(rec, uploadRequest(t, map[string]string{"folder": "projects"}, "x.png", pngBytes))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("upload with R2 off: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.Settings(rec, httptest.NewRequest(http.MethodGet, "/v1/admin/uploads", nil))
	if decodeBody(t, rec)["enabled"] != false {
		t.Errorf("settings with R2 off: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	uploadHandler(&memStore{}, 5<<20).Settings(rec, httptest.NewRequest(http.MethodGet, "/v1/admin/uploads", nil))
	got := decodeBody(t, rec)
	if got["enabled"] != true || got["maxBytes"] != float64(5<<20) {
		t.Errorf("settings with R2 on: %s", rec.Body.String())
	}
}
