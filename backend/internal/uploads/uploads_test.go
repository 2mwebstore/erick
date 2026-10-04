package uploads

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

// The smallest byte sequences each format is recognised by.
var samples = map[string][]byte{
	"image/png":  []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"),
	"image/jpeg": []byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00"),
	"image/gif":  []byte("GIF89a\x01\x00\x01\x00"),
	"image/webp": []byte("RIFF\x24\x00\x00\x00WEBPVP8 "),
	"image/avif": []byte("\x00\x00\x00\x1cftypavif\x00\x00\x00\x00"),
}

type memoryStore struct {
	key, contentType string
	body             []byte
	err              error
}

func (m *memoryStore) Put(_ context.Context, key, contentType string, body []byte) error {
	m.key, m.contentType, m.body = key, contentType, body
	return m.err
}

func uploader(store Store, max int64) *Uploader {
	u := NewUploader(store, "https://cdn.example.com/", max)
	u.now = func() time.Time { return time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC) }
	return u
}

func TestSniffRecognisesEachSupportedImage(t *testing.T) {
	for want, data := range samples {
		if got, _ := Sniff(data); got != want {
			t.Errorf("%s sniffed as %q", want, got)
		}
	}
}

// What a file is comes from its bytes: these are refused whatever they are named.
func TestSniffRefusesEverythingElse(t *testing.T) {
	for name, data := range map[string]string{
		"SVG, which can carry script": `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`,
		"an HTML page":                "<!doctype html><html><script>alert(1)</script></html>",
		"a PDF":                       "%PDF-1.7\n",
		"plain text":                  "just some text",
		"a ZIP":                       "PK\x03\x04",
	} {
		if got, _ := Sniff([]byte(data)); got != "" {
			t.Errorf("%s was taken for %s", name, got)
		}
	}
}

func TestUploadStoresUnderARandomKeyAndReturnsThePublicURL(t *testing.T) {
	store := &memoryStore{}
	img, err := uploader(store, 1<<20).Upload(context.Background(), "projects", strings.NewReader(string(samples["image/png"])))
	if err != nil {
		t.Fatal(err)
	}

	if !regexp.MustCompile(`^projects/2026/10/[0-9a-f]{32}\.png$`).MatchString(img.Key) {
		t.Errorf("key %q", img.Key)
	}
	if img.URL != "https://cdn.example.com/"+img.Key {
		t.Errorf("url %q", img.URL)
	}
	if store.key != img.Key || store.contentType != "image/png" || string(store.body) != string(samples["image/png"]) {
		t.Errorf("stored %q %q (%d bytes)", store.key, store.contentType, len(store.body))
	}

	second, _ := uploader(store, 1<<20).Upload(context.Background(), "projects", strings.NewReader(string(samples["image/png"])))
	if second.Key == img.Key {
		t.Error("two uploads got the same key")
	}
}

func TestUploadRefusals(t *testing.T) {
	cases := map[string]struct {
		folder string
		body   string
		want   error
	}{
		"an unknown folder":     {"../secrets", string(samples["image/png"]), ErrFolder},
		"a file over the limit": {"projects", string(samples["image/png"]) + strings.Repeat("x", 64), ErrTooLarge},
		"an empty file":         {"projects", "", ErrEmpty},
		"an HTML page":          {"projects", "<html><script>x</script>", ErrNotImage},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			store := &memoryStore{}
			_, err := uploader(store, 32).Upload(context.Background(), tc.folder, strings.NewReader(tc.body))
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if store.key != "" {
				t.Error("a refused file was stored")
			}
		})
	}
}

func TestUploadReportsAStoreFailure(t *testing.T) {
	_, err := uploader(&memoryStore{err: errors.New("bucket unreachable")}, 1<<20).
		Upload(context.Background(), "portraits", strings.NewReader(string(samples["image/jpeg"])))
	if err == nil || !strings.Contains(err.Error(), "bucket unreachable") {
		t.Errorf("got %v", err)
	}
}

// The R2 client against a stand-in S3 server: what actually goes over the wire.
func TestR2StoreSendsASignedPutWithTheRightHeaders(t *testing.T) {
	type seen struct {
		method, path, contentType, cacheControl, auth, encoding, trailer string
		body                                                             []byte
	}
	got := make(chan seen, 1)
	s3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got <- seen{
			method: r.Method, path: r.URL.Path,
			contentType: r.Header.Get("Content-Type"), cacheControl: r.Header.Get("Cache-Control"),
			auth: r.Header.Get("Authorization"), encoding: r.Header.Get("Content-Encoding"),
			trailer: r.Header.Get("X-Amz-Trailer"), body: body,
		}
		w.Header().Set("ETag", `"abc"`)
	}))
	t.Cleanup(s3.Close)

	store := NewR2Store(s3.URL, "key-id", "secret", "portfolio", nil)
	if err := store.Put(context.Background(), "projects/2026/10/x.png", "image/png", samples["image/png"]); err != nil {
		t.Fatal(err)
	}

	r := <-got
	if r.method != http.MethodPut || r.path != "/portfolio/projects/2026/10/x.png" {
		t.Errorf("%s %s — want a path-style PUT into the bucket", r.method, r.path)
	}
	if r.contentType != "image/png" {
		t.Errorf("content type %q", r.contentType)
	}
	if r.cacheControl != "public, max-age=31536000, immutable" {
		t.Errorf("cache control %q", r.cacheControl)
	}
	if !strings.HasPrefix(r.auth, "AWS4-HMAC-SHA256 Credential=key-id/") || !strings.Contains(r.auth, "/auto/s3/") {
		t.Errorf("authorization %q — want SigV4 for region auto", r.auth)
	}
	if strings.Contains(r.encoding, "aws-chunked") || r.trailer != "" {
		t.Errorf("checksum trailers were sent (encoding %q, trailer %q)", r.encoding, r.trailer)
	}
	if string(r.body) != string(samples["image/png"]) {
		t.Error("the body arrived altered")
	}
}
