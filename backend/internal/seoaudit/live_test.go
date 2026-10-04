package seoaudit

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLiveSite audits a real site and prints the report. It only runs when asked:
//
//	LIVE_SEO_AUDIT_URL=https://kongchansila.com go test ./internal/seoaudit -run LiveSite -v
func TestLiveSite(t *testing.T) {
	origin := os.Getenv("LIVE_SEO_AUDIT_URL")
	if origin == "" {
		t.Skip("set LIVE_SEO_AUDIT_URL to audit a real site")
	}

	a, err := New(origin, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	r, err := a.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s — score %d, sitelinks %d, %d pages in %d ms", r.Origin, r.Score, r.SitelinksScore, len(r.Pages), r.DurationMs)
	for _, c := range r.Checks {
		t.Logf("[%-4s] %-15s %s", c.Status, c.ID, c.Detail)
		for i, u := range c.URLs {
			if i == 5 {
				t.Logf("         … and %d more", len(c.URLs)-5)
				break
			}
			t.Logf("         %s", u)
		}
	}
}
