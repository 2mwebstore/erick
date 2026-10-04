// Package uploads stores images for the admin panel in Cloudflare R2 (any
// S3-compatible bucket works), so an image field can hold an uploaded file as
// well as a pasted link.
//
// What a file is, is decided from its bytes, never from the name or type the
// browser sent: an HTML page renamed photo.jpg is refused. SVG is refused too,
// because an SVG can carry script. The stored name is random and the caller's
// file name is never used, so nothing about the object key comes from the
// request.
package uploads

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	ErrFolder   = errors.New("uploads: unknown folder")
	ErrNotImage = errors.New("uploads: not a supported image")
	ErrTooLarge = errors.New("uploads: file too large")
	ErrEmpty    = errors.New("uploads: empty file")
)

// Folders are the places an image can be uploaded to, one per kind of field.
var Folders = map[string]bool{
	"projects":  true,
	"portraits": true,
	"profiles":  true,
}

// Types lists what is accepted, for messages and the admin panel.
var Types = []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/avif"}

// Store puts objects in the bucket and removes them.
type Store interface {
	Put(ctx context.Context, key, contentType string, body []byte) error
	Delete(ctx context.Context, key string) error
}

// ownKey matches exactly the keys Upload creates. Only those can be deleted:
// a file put in the bucket any other way — by hand, or by another app sharing
// the bucket — never matches, so it can never be removed from here.
var ownKey = regexp.MustCompile(`^(projects|portraits|profiles)/\d{4}/\d{2}/[0-9a-f]{32}\.(jpg|png|webp|gif|avif)$`)

// Image is a stored upload.
type Image struct {
	URL         string `json:"url"`
	Key         string `json:"key"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
}

// Uploader checks images and stores them.
type Uploader struct {
	store     Store
	publicURL string
	maxBytes  int64
	now       func() time.Time
}

// NewUploader stores into store and answers with URLs under publicURL, the
// bucket's public address (an r2.dev URL or a custom domain).
func NewUploader(store Store, publicURL string, maxBytes int64) *Uploader {
	return &Uploader{
		store:     store,
		publicURL: strings.TrimRight(publicURL, "/"),
		maxBytes:  maxBytes,
		now:       time.Now,
	}
}

// MaxBytes is the largest file accepted.
func (u *Uploader) MaxBytes() int64 { return u.maxBytes }

// Upload checks the image read from r and stores it under folder.
func (u *Uploader) Upload(ctx context.Context, folder string, r io.Reader) (*Image, error) {
	if !Folders[folder] {
		return nil, ErrFolder
	}

	data, err := io.ReadAll(io.LimitReader(r, u.maxBytes+1))
	if err != nil {
		return nil, err
	}
	switch {
	case int64(len(data)) > u.maxBytes:
		return nil, ErrTooLarge
	case len(data) == 0:
		return nil, ErrEmpty
	}

	contentType, ext := Sniff(data)
	if contentType == "" {
		return nil, ErrNotImage
	}

	name, err := randomName()
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s/%s/%s%s", folder, u.now().UTC().Format("2006/01"), name, ext)

	if err := u.store.Put(ctx, key, contentType, data); err != nil {
		return nil, fmt.Errorf("uploads: storing %s: %w", key, err)
	}

	return &Image{
		URL:         u.publicURL + "/" + key,
		Key:         key,
		ContentType: contentType,
		Size:        len(data),
	}, nil
}

// KeyFor returns the object key behind url when url is an image this uploader
// stored, and false for anything else: a pasted link, a file in public/, or an
// object in the bucket that Upload did not create.
func (u *Uploader) KeyFor(url string) (string, bool) {
	key, found := strings.CutPrefix(strings.TrimSpace(url), u.publicURL+"/")
	if !found || !ownKey.MatchString(key) {
		return "", false
	}
	return key, true
}

// Delete removes an image this uploader stored. A key that Upload would not
// have created is refused without touching the bucket.
func (u *Uploader) Delete(ctx context.Context, key string) error {
	if !ownKey.MatchString(key) {
		return fmt.Errorf("uploads: refusing to delete %q, which this site did not upload", key)
	}
	return u.store.Delete(ctx, key)
}

// Sniff names the image type from the file's first bytes, and the extension
// to store it with. Anything else answers "".
func Sniff(data []byte) (contentType, ext string) {
	// AVIF first: the standard library's sniffer does not know it.
	if len(data) >= 12 && string(data[4:8]) == "ftyp" {
		switch string(data[8:12]) {
		case "avif", "avis":
			return "image/avif", ".avif"
		}
	}
	switch http.DetectContentType(data) {
	case "image/jpeg":
		return "image/jpeg", ".jpg"
	case "image/png":
		return "image/png", ".png"
	case "image/webp":
		return "image/webp", ".webp"
	case "image/gif":
		return "image/gif", ".gif"
	}
	return "", ""
}

func randomName() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ── R2 ──────────────────────────────────────────────────────────────────────

// R2Store is a bucket reached through R2's S3-compatible API.
type R2Store struct {
	client *s3.Client
	bucket string
}

// NewR2Store connects to endpoint, such as https://<account>.r2.cloudflarestorage.com.
// A nil httpClient uses the SDK's default.
func NewR2Store(endpoint, accessKeyID, secretAccessKey, bucket string, httpClient *http.Client) *R2Store {
	opts := s3.Options{
		// R2 has no AWS regions; "auto" is what Cloudflare documents.
		Region:       "auto",
		BaseEndpoint: aws.String(endpoint),
		UsePathStyle: true,
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		// Newer SDK versions add checksum trailers to every upload by default,
		// which not every S3-compatible store accepts; Cloudflare's own R2
		// examples turn this off. Checksums are still sent where an operation
		// requires one.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}
	if httpClient != nil {
		opts.HTTPClient = httpClient
	}
	return &R2Store{client: s3.New(opts), bucket: bucket}
}

// Put stores body under key. Keys are never reused — each upload gets a random
// name — so the object can be cached for a year.
func (s *R2Store) Put(ctx context.Context, key, contentType string, body []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(body),
		ContentLength: aws.Int64(int64(len(body))),
		ContentType:   aws.String(contentType),
		CacheControl:  aws.String("public, max-age=31536000, immutable"),
	})
	return err
}

// Delete removes key. Deleting a key that is already gone succeeds.
func (s *R2Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}
