package seoaudit

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type route struct {
	status   int
	location string
	body     string
	ctype    string
}

// site serves routes, with {{origin}} replaced by the server's own address.
func site(t *testing.T, routes map[string]route) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rt, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		fill := func(s string) string { return strings.ReplaceAll(s, "{{origin}}", srv.URL) }
		if rt.location != "" {
			w.Header().Set("Location", fill(rt.location))
		}
		ctype := rt.ctype
		if ctype == "" {
			ctype = "text/html; charset=utf-8"
		}
		w.Header().Set("Content-Type", ctype)
		status := rt.status
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(fill(rt.body)))
	}))
	t.Cleanup(srv.Close)
	return srv
}

type pageSpec struct {
	title, description, canonical string
	hreflang                      map[string]string
	jsonLD                        []string
	nav, links                    []string
	h1                            int
	images                        string
}

func renderPage(p pageSpec) string {
	var head strings.Builder
	if p.title != "" {
		fmt.Fprintf(&head, "<title>%s</title>", p.title)
	}
	if p.description != "" {
		fmt.Fprintf(&head, `<meta name="description" content="%s">`, p.description)
	}
	if p.canonical != "" {
		fmt.Fprintf(&head, `<link rel="canonical" href="{{origin}}%s">`, p.canonical)
	}
	for lang, path := range p.hreflang {
		fmt.Fprintf(&head, `<link rel="alternate" hreflang="%s" href="{{origin}}%s">`, lang, path)
	}
	for _, ld := range p.jsonLD {
		fmt.Fprintf(&head, `<script type="application/ld+json">%s</script>`, ld)
	}

	var body strings.Builder
	if len(p.nav) > 0 {
		body.WriteString("<nav>")
		for _, href := range p.nav {
			fmt.Fprintf(&body, `<a href="%s">link</a>`, href)
		}
		body.WriteString("</nav>")
	}
	h1 := p.h1
	if h1 == 0 {
		h1 = 1
	} else if h1 < 0 {
		h1 = 0
	}
	for i := 0; i < h1; i++ {
		body.WriteString("<h1>Heading</h1>")
	}
	for _, href := range p.links {
		fmt.Fprintf(&body, `<a href="%s">link</a>`, href)
	}
	body.WriteString(p.images)
	return "<!doctype html><html><head>" + head.String() + "</head><body>" + body.String() + "</body></html>"
}

const (
	websiteLD    = `{"@context":"https://schema.org","@graph":[{"@type":"Person"},{"@type":"WebSite"}]}`
	breadcrumbLD = `{"@context":"https://schema.org","@type":"BreadcrumbList","itemListElement":[]}`
	robotsOK     = "User-agent: *\nAllow: /\nDisallow: /api/\n\nSitemap: {{origin}}/sitemap.xml\n"
)

func sitemap(paths ...string) route {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, p := range paths {
		fmt.Fprintf(&b, "<url><loc>{{origin}}%s</loc></url>", p)
	}
	b.WriteString("</urlset>")
	return route{body: b.String(), ctype: "application/xml"}
}

func homes() map[string]string {
	return map[string]string{"en": "/", "km": "/km", "x-default": "/"}
}

func run(t *testing.T, origin string) *Report {
	t.Helper()
	a, err := New(origin, &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	report, err := a.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func statusOf(t *testing.T, r *Report, id string) Check {
	t.Helper()
	for _, c := range r.Checks {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("no check %q", id)
	return Check{}
}

func TestHealthySitePassesEveryCheck(t *testing.T) {
	srv := site(t, map[string]route{
		"/robots.txt":  {body: robotsOK, ctype: "text/plain"},
		"/sitemap.xml": sitemap("/", "/km", "/work", "/work/a", "/work/b", "/resume"),
		"/": {body: renderPage(pageSpec{title: "Home", description: "Home page", canonical: "/", hreflang: homes(),
			jsonLD: []string{websiteLD}, nav: []string{"#about", "/work", "/resume", "/km"},
			images: `<img src="/p.png" alt="Portrait"><img src="/d.svg" alt="">`})},
		"/km":     {body: renderPage(pageSpec{title: "Home km", description: "Home page km", canonical: "/km", hreflang: homes(), nav: []string{"/work", "/resume", "/"}})},
		"/work":   {body: renderPage(pageSpec{title: "Work", description: "All work", canonical: "/work", nav: []string{"/", "/km"}, links: []string{"/work/a", "/work/b"}})},
		"/work/a": {body: renderPage(pageSpec{title: "A", description: "Project A", canonical: "/work/a", jsonLD: []string{breadcrumbLD}, links: []string{"/work", "/work/b", "/resume"}})},
		"/work/b": {body: renderPage(pageSpec{title: "B", description: "Project B", canonical: "/work/b", jsonLD: []string{breadcrumbLD}, links: []string{"/work", "/work/a", "/km"}})},
		"/resume": {body: renderPage(pageSpec{title: "Resume", description: "Resume", canonical: "/resume", links: []string{"/"}})},
	})

	r := run(t, srv.URL)

	for _, c := range r.Checks {
		want := Pass
		if c.ID == "https" {
			want = Warn // the test server is plain HTTP
		}
		if c.Status != want {
			t.Errorf("%s: %s — %s %v", c.ID, c.Status, c.Detail, c.URLs)
		}
	}
	if r.SitelinksScore != 100 {
		t.Errorf("sitelinks score %d, want 100", r.SitelinksScore)
	}
	if len(r.Pages) != 6 {
		t.Errorf("crawled %d pages, want 6", len(r.Pages))
	}
	for _, p := range r.Pages {
		if p.Depth < 0 || !p.InSitemap {
			t.Errorf("%s: depth %d, inSitemap %v", p.URL, p.Depth, p.InSitemap)
		}
	}
}

func TestBrokenSiteIsReportedCheckByCheck(t *testing.T) {
	var offsiteHits atomic.Int32
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offsiteHits.Add(1)
	}))
	t.Cleanup(other.Close)

	srv := site(t, map[string]route{
		"/robots.txt":  {body: "User-agent: *\nDisallow: /\n", ctype: "text/plain"},
		"/sitemap.xml": sitemap("/", "/km", "/work", "/work/a", "/work/b", "/resume", "/orphan"),
		"/": {body: renderPage(pageSpec{title: "Home", description: "Home page", canonical: "/", hreflang: homes(),
			jsonLD: []string{websiteLD}, nav: []string{"#about", "#work", "/"},
			links: []string{"/work", "/resume", "/km", "/old", "/missing", "/away", other.URL + "/elsewhere"}})},
		"/km":      {body: renderPage(pageSpec{title: "Home km", description: "Home page km", canonical: "/km", hreflang: map[string]string{"km": "/km"}, links: []string{"/"}})},
		"/work":    {body: renderPage(pageSpec{title: "Same title", description: "All work", canonical: "/work", links: []string{"/work/a", "/work/b"}, images: `<img src="/x.png">`})},
		"/work/a":  {body: renderPage(pageSpec{title: "A", description: "Project A", canonical: "/work/a", jsonLD: []string{`{"@type": broken`}, links: []string{"/work"}})},
		"/work/b":  {body: renderPage(pageSpec{title: "Same title", canonical: "/work/b", h1: 2, jsonLD: []string{breadcrumbLD}, links: []string{"/work"}})},
		"/resume":  {body: renderPage(pageSpec{title: "Resume", description: "Resume", canonical: "/", links: []string{"/"}})},
		"/orphan":  {body: renderPage(pageSpec{title: "Orphan", description: "Nobody links here", canonical: "/orphan"})},
		"/old":     {status: http.StatusMovedPermanently, location: "{{origin}}/work"},
		"/away":    {status: http.StatusFound, location: other.URL + "/landing"},
		"/missing": {status: http.StatusNotFound, body: "gone"},
	})

	r := run(t, srv.URL)

	want := map[string]Status{
		"navigation":      Warn, // only #section jumps in the <nav>
		"orphans":         Fail, // /orphan
		"weak_links":      Warn,
		"breadcrumbs":     Warn, // /work/a has no valid BreadcrumbList
		"robots":          Fail, // Disallow: /
		"sitemap":         Pass,
		"sitemap_pages":   Fail, // /resume canonical points at /
		"broken_links":    Fail, // /missing
		"redirect_links":  Warn, // /old
		"canonical":       Warn, // /resume names / as canonical
		"hreflang":        Warn, // /km does not link back to /
		"titles":          Warn, // "Same title" twice
		"descriptions":    Warn, // /work/b has none
		"h1":              Warn, // /work/b has two
		"structured_data": Fail, // invalid JSON on /work/a
		"image_alt":       Warn, // /work
	}
	for id, status := range want {
		if c := statusOf(t, r, id); c.Status != status {
			t.Errorf("%s: got %s, want %s — %s %v", id, c.Status, status, c.Detail, c.URLs)
		}
	}

	if !strings.Contains(strings.Join(statusOf(t, r, "orphans").URLs, " "), "/orphan") {
		t.Error("the orphan page is not named")
	}
	if !strings.Contains(strings.Join(statusOf(t, r, "broken_links").URLs, " "), "/missing ← linked from") {
		t.Error("the broken link does not say where it was found")
	}

	// The audit must never leave the configured site: not for a link, not for a redirect.
	if n := offsiteHits.Load(); n != 0 {
		t.Errorf("the audit made %d request(s) to another server", n)
	}
	if r.Score >= 90 {
		t.Errorf("a site this broken scored %d", r.Score)
	}
}

// The same description on a page and its own translation is an untranslated
// field, reported as such rather than as two pages competing.
func TestTextSharedByLanguageVersionsIsReportedAsUntranslated(t *testing.T) {
	pair := map[string]string{"en": "/work", "km": "/km/work"}
	srv := site(t, map[string]route{
		"/robots.txt":  {body: robotsOK, ctype: "text/plain"},
		"/sitemap.xml": sitemap("/", "/work", "/km/work", "/other"),
		"/":            {body: renderPage(pageSpec{title: "Home", description: "Home", canonical: "/", links: []string{"/work", "/km/work", "/other"}})},
		"/work":        {body: renderPage(pageSpec{title: "Work", description: "Shared text", canonical: "/work", hreflang: pair})},
		"/km/work":     {body: renderPage(pageSpec{title: "Work km", description: "Shared text", canonical: "/km/work", hreflang: pair})},
		"/other":       {body: renderPage(pageSpec{title: "Other", description: "Other", canonical: "/other"})},
	})

	c := statusOf(t, run(t, srv.URL), "descriptions")
	if c.Status != Warn || !strings.Contains(c.Detail, "Translate it") {
		t.Errorf("got %s — %s", c.Status, c.Detail)
	}
}

func TestSitemapProblems(t *testing.T) {
	cases := map[string]route{
		"not XML":        {body: "<urlset><url>", ctype: "application/xml"},
		"wrong root":     {body: `<?xml version="1.0"?><urls xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"></urls>`, ctype: "application/xml"},
		"another site":   {body: `<?xml version="1.0"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://elsewhere.example/</loc></url><url><loc>{{origin}}/</loc></url></urlset>`, ctype: "application/xml"},
		"listed twice":   {body: `<?xml version="1.0"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>{{origin}}/</loc></url><url><loc>{{origin}}</loc></url></urlset>`, ctype: "application/xml"},
		"missing (404)":  {status: http.StatusNotFound},
		"server failure": {status: http.StatusInternalServerError},
	}
	for name, sm := range cases {
		t.Run(name, func(t *testing.T) {
			srv := site(t, map[string]route{
				"/robots.txt":  {body: robotsOK, ctype: "text/plain"},
				"/sitemap.xml": sm,
				"/":            {body: renderPage(pageSpec{title: "Home", description: "Home", canonical: "/"})},
			})
			if c := statusOf(t, run(t, srv.URL), "sitemap"); c.Status != Fail {
				t.Errorf("got %s — %s", c.Status, c.Detail)
			}
		})
	}
}

func TestSitemapIndexIsFollowed(t *testing.T) {
	srv := site(t, map[string]route{
		"/robots.txt": {body: robotsOK, ctype: "text/plain"},
		"/sitemap.xml": {ctype: "application/xml", body: `<?xml version="1.0"?><sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` +
			`<sitemap><loc>{{origin}}/sitemaps/pages.xml</loc></sitemap></sitemapindex>`},
		"/sitemaps/pages.xml": sitemap("/"),
		"/":                   {body: renderPage(pageSpec{title: "Home", description: "Home", canonical: "/", jsonLD: []string{websiteLD}})},
	})
	if c := statusOf(t, run(t, srv.URL), "sitemap"); c.Status != Pass {
		t.Errorf("got %s — %s %v", c.Status, c.Detail, c.URLs)
	}
}

func TestParseRobots(t *testing.T) {
	blocks, sitemaps := parseRobots("User-agent: Googlebot\nDisallow: /\n\nUser-agent: *\nAllow: /\n# Disallow: /\nSitemap: https://example.com/sitemap.xml\n")
	if blocks {
		t.Error("a Googlebot-only block, or a commented rule, was read as blocking everyone")
	}
	if len(sitemaps) != 1 {
		t.Errorf("sitemaps %v", sitemaps)
	}

	if blocks, _ := parseRobots("User-agent: *\nDisallow: /\n"); !blocks {
		t.Error("Disallow: / for * was not detected")
	}
	if blocks, _ := parseRobots("User-agent: *\nDisallow:\n"); blocks {
		t.Error("an empty Disallow allows everything")
	}
}

func TestLinksAreNormalisedToPagesOnThisSite(t *testing.T) {
	a, err := New("https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := url.Parse("https://example.com/work/a")

	cases := map[string]string{
		"/work":                       "https://example.com/work",
		"/work#top":                   "https://example.com/work",
		"/#contact":                   "https://example.com/",
		"../resume":                   "https://example.com/resume",
		"https://example.com":         "https://example.com/",
		"#section":                    "",
		"mailto:me@example.com":       "",
		"https://other.example/work":  "",
		"http://example.com/work":     "", // another scheme is another origin
		"/_nuxt/entry.js":             "",
		"/api/content":                "",
		"/og-default.png":             "",
		"/cdn-cgi/l/email-protection": "",
	}
	for href, want := range cases {
		if got := a.resolve(base, href); got != want {
			t.Errorf("%q: got %q, want %q", href, got, want)
		}
	}
}

func TestNewRejectsAnythingButAnAbsoluteOrigin(t *testing.T) {
	for _, origin := range []string{"", "example.com", "/relative", "ftp://example.com", "javascript:alert(1)"} {
		if _, err := New(origin, nil); err == nil {
			t.Errorf("%q was accepted", origin)
		}
	}
	a, err := New("https://example.com/some/path?q=1", nil)
	if err != nil || a.Origin() != "https://example.com" {
		t.Errorf("origin %q, err %v", a.Origin(), err)
	}
}
