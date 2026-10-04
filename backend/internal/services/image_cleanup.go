package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/models"
	"github.com/kongchansila/portfolio/backend/internal/uploads"
)

// ImageRefs answers whether an image address is still shown anywhere.
type ImageRefs interface {
	ImageInUse(ctx context.Context, url string) (bool, error)
}

// ImageCleanup deletes uploaded images from the bucket once nothing shows them:
// after a project is deleted, or an image field is replaced or cleared.
//
// An image is deleted only when all of these hold:
//   - it is a file this site uploaded (uploads.Uploader.KeyFor) — a pasted link,
//     a file in public/, or another app's file in a shared bucket is never touched;
//   - no project, portrait or profile logo still uses it — the same file can be
//     used in two places;
//   - that check ran: if it fails, the file is kept.
//
// It runs after the save has succeeded, so a failed save never deletes a file,
// and in the background, so a slow or failing bucket never fails a save.
type ImageCleanup struct {
	uploader *uploads.Uploader
	refs     ImageRefs
	audit    *AuditService
	log      *slog.Logger

	wg sync.WaitGroup
}

// NewImageCleanup returns nil when uploads are off; a nil cleanup does nothing.
func NewImageCleanup(uploader *uploads.Uploader, refs ImageRefs, audit *AuditService, log *slog.Logger) *ImageCleanup {
	if uploader == nil || refs == nil {
		return nil
	}
	return &ImageCleanup{uploader: uploader, refs: refs, audit: audit, log: log}
}

// Release deletes, in the background, each of urls that this site uploaded
// and that nothing uses any more. reason is recorded in the audit log.
func (c *ImageCleanup) Release(user *models.User, ip, reason string, urls ...string) {
	if c == nil || len(urls) == 0 {
		return
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		// A panic here would take the API down: request recovery does not
		// cover background goroutines.
		defer func() {
			if p := recover(); p != nil {
				c.log.Error("image cleanup stopped unexpectedly", slog.String("panic", fmt.Sprint(p)))
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		c.ReleaseNow(ctx, user, ip, reason, urls...)
	}()
}

// ReleaseNow is Release, done before returning.
func (c *ImageCleanup) ReleaseNow(ctx context.Context, user *models.User, ip, reason string, urls ...string) {
	if c == nil {
		return
	}
	seen := map[string]bool{}
	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" || seen[url] {
			continue
		}
		seen[url] = true

		key, ours := c.uploader.KeyFor(url)
		if !ours {
			continue
		}

		inUse, err := c.refs.ImageInUse(ctx, url)
		if err != nil {
			c.log.Warn("image kept: could not check whether it is still used",
				slog.String("key", key), slog.String("error", err.Error()))
			continue
		}
		if inUse {
			c.log.Info("image kept: still used elsewhere", slog.String("key", key))
			continue
		}

		if err := c.uploader.Delete(ctx, key); err != nil {
			c.log.Error("deleting image from the bucket", slog.String("key", key), slog.String("error", err.Error()))
			continue
		}
		c.log.Info("image deleted from the bucket", slog.String("key", key), slog.String("reason", reason))
		if c.audit != nil {
			c.audit.Record(ctx, user, "delete", "image", key, map[string]any{"reason": reason}, ip)
		}
	}
}

// Wait blocks until background deletions have finished, for a clean shutdown.
func (c *ImageCleanup) Wait() {
	if c != nil {
		c.wg.Wait()
	}
}

// RemovedSettingImages lists the image addresses a settings save stops using:
// the old portrait when it changes, and logos dropped from the profiles.
// Only keys present in next are considered, since a save may change a few
// settings and leave the rest as they were.
func RemovedSettingImages(previous, next map[string]string) []string {
	var removed []string
	if v, ok := next["portrait"]; ok {
		if old := strings.TrimSpace(previous["portrait"]); old != "" && old != strings.TrimSpace(v) {
			removed = append(removed, old)
		}
	}
	if v, ok := next["profiles"]; ok {
		kept := map[string]bool{}
		for _, img := range profileImages(v) {
			kept[img] = true
		}
		for _, img := range profileImages(previous["profiles"]) {
			if !kept[img] {
				removed = append(removed, img)
			}
		}
	}
	return removed
}

func profileImages(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var profiles []struct {
		Image string `json:"image"`
	}
	if err := json.Unmarshal([]byte(raw), &profiles); err != nil {
		return nil
	}
	var out []string
	for _, p := range profiles {
		if img := strings.TrimSpace(p.Image); img != "" {
			out = append(out, img)
		}
	}
	return out
}
