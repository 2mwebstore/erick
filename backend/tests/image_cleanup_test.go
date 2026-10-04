package tests

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"

	"github.com/kongchansila/portfolio/backend/internal/services"
	"github.com/kongchansila/portfolio/backend/internal/uploads"
)

const cdn = "https://cdn.example.com/"

// Keys shaped exactly like the ones Upload creates.
const (
	oldPhoto  = "projects/2026/10/00000000000000000000000000000001.png"
	newPhoto  = "projects/2026/10/00000000000000000000000000000002.png"
	shared    = "portraits/2026/10/00000000000000000000000000000003.webp"
	otherApps = "products/1696000000-abc12345.jpg"
)

type recordingStore struct {
	mu      sync.Mutex
	deleted []string
	fail    error
}

func (s *recordingStore) Put(context.Context, string, string, []byte) error { return nil }
func (s *recordingStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleted = append(s.deleted, key)
	return s.fail
}
func (s *recordingStore) keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := append([]string(nil), s.deleted...)
	sort.Strings(out)
	return out
}

type fakeRefs struct {
	inUse map[string]bool
	err   error
}

func (f fakeRefs) ImageInUse(_ context.Context, url string) (bool, error) {
	return f.inUse[url], f.err
}

func cleanup(store *recordingStore, refs services.ImageRefs) *services.ImageCleanup {
	return services.NewImageCleanup(uploads.NewUploader(store, cdn, 1<<20), refs, nil, quietLog())
}

func TestCleanupDeletesAnUploadNothingUses(t *testing.T) {
	store := &recordingStore{}
	cleanup(store, fakeRefs{}).ReleaseNow(context.Background(), nil, "", "test", cdn+oldPhoto)

	if got := store.keys(); len(got) != 1 || got[0] != oldPhoto {
		t.Errorf("deleted %v", got)
	}
}

// The same file can be shown in two places; deleting it with one would break the other.
func TestCleanupKeepsAnUploadStillUsedElsewhere(t *testing.T) {
	store := &recordingStore{}
	cleanup(store, fakeRefs{inUse: map[string]bool{cdn + shared: true}}).
		ReleaseNow(context.Background(), nil, "", "test", cdn+shared)

	if got := store.keys(); len(got) != 0 {
		t.Errorf("a file still in use was deleted: %v", got)
	}
}

// When in doubt, keep it: an image that cannot be checked is not deleted.
func TestCleanupKeepsAnUploadWhenTheCheckFails(t *testing.T) {
	store := &recordingStore{}
	cleanup(store, fakeRefs{err: errors.New("database down")}).
		ReleaseNow(context.Background(), nil, "", "test", cdn+oldPhoto)

	if got := store.keys(); len(got) != 0 {
		t.Errorf("deleted without a successful check: %v", got)
	}
}

func TestCleanupNeverTouchesFilesItDidNotUpload(t *testing.T) {
	store := &recordingStore{}
	cleanup(store, fakeRefs{}).ReleaseNow(context.Background(), nil, "", "test",
		"https://i.imgur.com/8HbEl4Q.png", // a pasted link
		"/portrait.jpg",                   // a file in public/
		cdn+otherApps,                     // another app's file in a shared bucket
		"",
	)
	if got := store.keys(); len(got) != 0 {
		t.Errorf("deleted %v", got)
	}
}

func TestCleanupDeletesEachFileOnceAndRunsInTheBackground(t *testing.T) {
	store := &recordingStore{}
	c := cleanup(store, fakeRefs{})
	c.Release(nil, "", "test", cdn+oldPhoto, cdn+oldPhoto, " "+cdn+newPhoto+" ")
	c.Wait()

	if got := store.keys(); len(got) != 2 || got[0] != oldPhoto || got[1] != newPhoto {
		t.Errorf("deleted %v", got)
	}
}

func TestCleanupIsOffWithoutUploads(t *testing.T) {
	c := services.NewImageCleanup(nil, fakeRefs{}, nil, quietLog())
	if c != nil {
		t.Fatal("expected no cleanup when uploads are off")
	}
	// A nil cleanup is safe to call: handlers do not check.
	c.Release(nil, "", "test", cdn+oldPhoto)
	c.Wait()
}

func TestRemovedSettingImages(t *testing.T) {
	previous := map[string]string{
		"portrait": cdn + oldPhoto,
		"profiles": `[{"image":"` + cdn + shared + `"},{"image":"https://i.imgur.com/x.png"},{"image":""}]`,
	}

	cases := map[string]struct {
		next map[string]string
		want []string
	}{
		"portrait replaced":          {map[string]string{"portrait": cdn + newPhoto}, []string{cdn + oldPhoto}},
		"portrait cleared":           {map[string]string{"portrait": ""}, []string{cdn + oldPhoto}},
		"portrait unchanged":         {map[string]string{"portrait": cdn + oldPhoto}, nil},
		"portrait not in this save":  {map[string]string{"name": "Someone"}, nil},
		"a logo removed":             {map[string]string{"profiles": `[{"image":"https://i.imgur.com/x.png"}]`}, []string{cdn + shared}},
		"logos reordered, none gone": {map[string]string{"profiles": `[{"image":"https://i.imgur.com/x.png"},{"image":"` + cdn + shared + `"}]`}, nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := services.RemovedSettingImages(previous, tc.next)
			sort.Strings(got)
			if len(got) != len(tc.want) || (len(got) > 0 && got[0] != tc.want[0]) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
