// Package seoaudit crawls the public site the way a search engine does and
// reports what would stop one from understanding it: pages no link leads to,
// links to missing or moved pages, canonical and hreflang mistakes, and missing
// titles, descriptions or structured data.
//
// It only ever requests the configured origin. That origin comes from
// configuration, never from a request, so the endpoint that runs an audit cannot
// be pointed at another server.
//
// The score it produces is this site's own diagnostic. It is not anything
// Google reports, and no score makes Google show sitelinks.
package seoaudit

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	// Google's own limits: 50,000 URLs and 50 MB per sitemap file.
	sitemapMaxURLs  = 50000
	sitemapMaxBytes = 50 << 20

	pageMaxBytes    = 5 << 20
	defaultMaxPages = 200
	userAgent       = "PortfolioSEOCheck/1.0 (site owner's own audit)"
)

// Status of one check.
type Status string

const (
	Pass Status = "pass"
	Warn Status = "warn"
	Fail Status = "fail"
)

// Check groups, in the order they are shown.
const (
	GroupSitelinks = "sitelinks"
	GroupTechnical = "technical"
	GroupOnPage    = "onpage"
)

// Check is one finding.
type Check struct {
	ID     string   `json:"id"`
	Group  string   `json:"group"`
	Title  string   `json:"title"`
	Status Status   `json:"status"`
	Detail string   `json:"detail"`
	URLs   []string `json:"urls,omitempty"`
}

// Page is what the crawl learned about one URL.
type Page struct {
	URL              string            `json:"url"`
	Status           int               `json:"status"`
	Error            string            `json:"error,omitempty"`
	RedirectTo       string            `json:"redirectTo,omitempty"`
	Depth            int               `json:"depth"` // clicks from the home page; -1 when no link leads here
	InSitemap        bool              `json:"inSitemap"`
	Indexable        bool              `json:"indexable"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Canonical        string            `json:"canonical"`
	H1               int               `json:"h1"`
	ImagesWithoutAlt int               `json:"imagesWithoutAlt"`
	StructuredData   []string          `json:"structuredData"`
	Hreflang         map[string]string `json:"hreflang,omitempty"`
	Incoming         int               `json:"incoming"`
	Outgoing         int               `json:"outgoing"`

	html           bool
	navTargets     []string
	links          []string
	invalidJSONLD  int
	redirectTarget string // RedirectTo when it stays on this site, else ""
}

// Report is the result of one audit.
type Report struct {
	Origin         string    `json:"origin"`
	StartedAt      time.Time `json:"startedAt"`
	DurationMs     int64     `json:"durationMs"`
	Score          int       `json:"score"`
	SitelinksScore int       `json:"sitelinksScore"`
	Truncated      bool      `json:"truncated"`
	Checks         []Check   `json:"checks"`
	Pages          []Page    `json:"pages"`
}

// Auditor crawls one origin.
type Auditor struct {
	origin   *url.URL
	client   *http.Client
	maxPages int
}

// New returns an auditor for origin, such as https://example.com. A nil client
// gets a default with a 15-second timeout.
func New(origin string, client *http.Client) (*Auditor, error) {
	u, err := url.Parse(strings.TrimSpace(origin))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("seoaudit: %q is not an absolute http(s) URL", origin)
	}
	u = &url.URL{Scheme: u.Scheme, Host: u.Host}

	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	c := *client
	// Redirects are reported, not followed: a link that lands on a redirect is
	// itself a finding.
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	return &Auditor{origin: u, client: &c, maxPages: defaultMaxPages}, nil
}

// Origin is the site this auditor crawls.
func (a *Auditor) Origin() string { return a.origin.String() }

// Run crawls the site and returns the report.
func (a *Auditor) Run(ctx context.Context) (*Report, error) {
	started := time.Now()
	home := a.origin.String() + "/"

	robots := a.checkRobots(ctx)
	sitemapURLs, sitemapCheck := a.readSitemap(ctx)
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	pages := map[string]*Page{}
	var order []string
	truncated := false

	visit := func(u string, depth int) *Page {
		p := a.visit(ctx, u)
		p.Depth = depth
		pages[u] = p
		order = append(order, u)
		return p
	}

	// Breadth first from the home page, so Depth is the fewest clicks to a page.
	depth := map[string]int{home: 0}
	queue := []string{home}
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(pages) >= a.maxPages {
			truncated = true
			break
		}
		u := queue[0]
		queue = queue[1:]

		p := visit(u, depth[u])
		next := p.links
		// Only a redirect that stays on this site is followed; one that leaves
		// it is reported, never fetched.
		if p.redirectTarget != "" {
			next = append([]string{p.redirectTarget}, next...)
		}
		for _, target := range next {
			if _, seen := depth[target]; !seen {
				step := 1
				if target == p.redirectTarget {
					step = 0 // a redirect is not a click
				}
				depth[target] = depth[u] + step
				queue = append(queue, target)
			}
		}
	}

	// Sitemap pages no link reached are fetched too, so their own state is known.
	for _, u := range sitemapURLs {
		if _, done := pages[u]; done {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(pages) >= a.maxPages {
			truncated = true
			break
		}
		visit(u, -1)
	}

	inSitemap := map[string]bool{}
	for _, u := range sitemapURLs {
		inSitemap[u] = true
	}

	incoming := map[string]map[string]bool{}
	for _, u := range order {
		p := pages[u]
		p.InSitemap = inSitemap[u]
		targets := map[string]bool{}
		for _, t := range p.links {
			if t == u {
				continue
			}
			targets[t] = true
			if incoming[t] == nil {
				incoming[t] = map[string]bool{}
			}
			incoming[t][u] = true
		}
		p.Outgoing = len(targets)
	}
	for _, u := range order {
		pages[u].Incoming = len(incoming[u])
	}

	s := &state{origin: a.origin, home: home, pages: pages, order: order, sitemap: sitemapURLs, incoming: incoming}
	checks := []Check{
		s.checkNavigation(),
		s.checkOrphans(),
		s.checkClickDepth(),
		s.checkWeakLinks(),
		s.checkBreadcrumbs(),
		a.checkHTTPS(ctx),
		robots,
		sitemapCheck,
		s.checkSitemapPages(),
		s.checkBrokenLinks(),
		s.checkRedirectLinks(),
		s.checkCanonicals(),
		s.checkHreflang(),
		s.checkTitles(),
		s.checkDescriptions(),
		s.checkHeadings(),
		s.checkStructuredData(),
		s.checkImageAlt(),
	}

	report := &Report{
		Origin:         a.origin.String(),
		StartedAt:      started.UTC(),
		DurationMs:     time.Since(started).Milliseconds(),
		Score:          score(checks, ""),
		SitelinksScore: score(checks, GroupSitelinks),
		Truncated:      truncated,
		Checks:         checks,
	}
	for _, u := range order {
		p := *pages[u]
		if p.StructuredData == nil {
			p.StructuredData = []string{}
		}
		report.Pages = append(report.Pages, p)
	}
	return report, nil
}

// score averages the checks of a group (or all of them for ""): a pass counts
// fully, a warning half.
func score(checks []Check, group string) int {
	var total, n float64
	for _, c := range checks {
		if group != "" && c.Group != group {
			continue
		}
		n++
		switch c.Status {
		case Pass:
			total++
		case Warn:
			total += 0.5
		}
	}
	if n == 0 {
		return 0
	}
	return int(total/n*100 + 0.5)
}

// ── Fetching ────────────────────────────────────────────────────────────────

type response struct {
	status   int
	header   http.Header
	body     []byte
	location string
}

func (a *Auditor) get(ctx context.Context, rawURL string, limit int64) (*response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	res, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil {
		return nil, err
	}

	out := &response{status: res.StatusCode, header: res.Header, body: body}
	if loc := res.Header.Get("Location"); loc != "" {
		if u, err := req.URL.Parse(loc); err == nil {
			out.location = u.String()
		}
	}
	return out, nil
}

func (a *Auditor) visit(ctx context.Context, u string) *Page {
	p := &Page{URL: u}

	res, err := a.get(ctx, u, pageMaxBytes)
	if err != nil {
		p.Error = err.Error()
		return p
	}
	p.Status = res.status

	if res.status >= 300 && res.status < 400 {
		p.RedirectTo = res.location
		if internal := a.internal(res.location); internal != "" {
			p.RedirectTo, p.redirectTarget = internal, internal
		}
		return p
	}

	p.html = strings.Contains(strings.ToLower(res.header.Get("Content-Type")), "text/html")
	noindex := strings.Contains(strings.ToLower(res.header.Get("X-Robots-Tag")), "noindex")
	if !p.html {
		p.Indexable = res.status == http.StatusOK && !noindex
		return p
	}

	base, _ := url.Parse(u)
	a.parse(base, res.body, p, &noindex)
	p.Indexable = res.status == http.StatusOK && !noindex
	return p
}

// internal normalises a link into the URL it would be crawled as, or returns ""
// when it leaves this origin or is not a page (API routes, build assets, files).
func (a *Auditor) internal(rawURL string) string {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return a.normalize(u)
}

func (a *Auditor) resolve(base *url.URL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") {
		return "" // a jump within the same page
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return a.normalize(base.ResolveReference(ref))
}

func (a *Auditor) normalize(u *url.URL) string {
	if skipPath(u.Path) {
		return ""
	}
	return a.onSite(u)
}

// onSite returns u in canonical form when it is on this origin, else "". Unlike
// normalize it accepts any path, for sitemap files and the URLs they list.
func (a *Auditor) onSite(u *url.URL) string {
	if u.Scheme != a.origin.Scheme || u.Host != a.origin.Host || u.User != nil {
		return ""
	}
	clean := &url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path, RawQuery: u.RawQuery}
	if clean.Path == "" {
		clean.Path = "/"
	}
	return clean.String()
}

func (a *Auditor) onSiteRaw(rawURL string) string {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return a.onSite(u)
}

func skipPath(p string) bool {
	for _, prefix := range []string{"/api/", "/_nuxt/", "/_ipx/", "/_i18n/", "/cdn-cgi/", "/__"} {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	switch strings.ToLower(path.Ext(p)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".avif", ".svg", ".ico",
		".pdf", ".xml", ".txt", ".json", ".css", ".js", ".woff", ".woff2", ".zip":
		return true
	}
	return false
}

// canon is the comparable form of an absolute URL: no fragment, and "/" for an
// empty path, so https://example.com and https://example.com/ are one page.
func canon(rawURL string) string {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" {
		return strings.TrimSpace(rawURL)
	}
	u.Fragment, u.RawFragment = "", ""
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String()
}

// ── HTML ────────────────────────────────────────────────────────────────────

func attr(n *html.Node, name string) (string, bool) {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val, true
		}
	}
	return "", false
}

func text(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

func (a *Auditor) parse(base *url.URL, body []byte, p *Page, noindex *bool) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return
	}

	seen := map[string]bool{}
	navSeen := map[string]bool{}
	titleFound := false

	var walk func(n *html.Node, inNav bool)
	walk = func(n *html.Node, inNav bool) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "nav":
				inNav = true
			case "title":
				if !titleFound {
					p.Title = text(n)
					titleFound = true
				}
			case "meta":
				name, _ := attr(n, "name")
				content, _ := attr(n, "content")
				switch strings.ToLower(name) {
				case "description":
					p.Description = strings.TrimSpace(content)
				case "robots", "googlebot":
					if strings.Contains(strings.ToLower(content), "noindex") {
						*noindex = true
					}
				}
			case "link":
				rel, _ := attr(n, "rel")
				href, _ := attr(n, "href")
				switch strings.ToLower(rel) {
				case "canonical":
					if p.Canonical == "" {
						if u, err := base.Parse(href); err == nil {
							p.Canonical = u.String()
						}
					}
				case "alternate":
					if lang, ok := attr(n, "hreflang"); ok {
						if u, err := base.Parse(href); err == nil {
							if p.Hreflang == nil {
								p.Hreflang = map[string]string{}
							}
							p.Hreflang[lang] = u.String()
						}
					}
				}
			case "h1":
				p.H1++
			case "img":
				if _, ok := attr(n, "alt"); !ok {
					p.ImagesWithoutAlt++
				}
			case "script":
				if typ, _ := attr(n, "type"); strings.EqualFold(typ, "application/ld+json") {
					// Raw script text: text() squeezes whitespace, which would
					// alter string values inside the JSON.
					raw := ""
					if n.FirstChild != nil {
						raw = n.FirstChild.Data
					}
					var data any
					if err := json.Unmarshal([]byte(raw), &data); err != nil {
						p.invalidJSONLD++
					} else {
						p.StructuredData = append(p.StructuredData, ldTypes(data)...)
					}
				}
			case "a":
				if href, ok := attr(n, "href"); ok {
					if target := a.resolve(base, href); target != "" {
						if !seen[target] {
							seen[target] = true
							p.links = append(p.links, target)
						}
						if inNav && !navSeen[target] {
							navSeen[target] = true
							p.navTargets = append(p.navTargets, target)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, inNav)
		}
	}
	walk(doc, false)
}

// ldTypes collects every @type in a JSON-LD document, @graph included.
func ldTypes(v any) []string {
	var out []string
	switch t := v.(type) {
	case map[string]any:
		switch typ := t["@type"].(type) {
		case string:
			out = append(out, typ)
		case []any:
			for _, x := range typ {
				if s, ok := x.(string); ok {
					out = append(out, s)
				}
			}
		}
		if graph, ok := t["@graph"]; ok {
			out = append(out, ldTypes(graph)...)
		}
	case []any:
		for _, x := range t {
			out = append(out, ldTypes(x)...)
		}
	}
	return out
}

// ── robots.txt, sitemap, HTTPS ──────────────────────────────────────────────

func (a *Auditor) checkRobots(ctx context.Context) Check {
	c := Check{ID: "robots", Group: GroupTechnical, Title: "robots.txt lets search engines in"}
	u := a.origin.String() + "/robots.txt"

	res, err := a.get(ctx, u, 512<<10)
	if err != nil {
		return c.with(Fail, "robots.txt could not be fetched: "+err.Error(), u)
	}
	if res.status == http.StatusNotFound {
		return c.with(Warn, "There is no robots.txt. Crawling is allowed, but nothing points crawlers at the sitemap.", u)
	}
	if res.status != http.StatusOK {
		return c.with(Fail, fmt.Sprintf("robots.txt answered %d. Google treats a 5xx here as \"crawl nothing\".", res.status), u)
	}

	blocksAll, sitemaps := parseRobots(string(res.body))
	if blocksAll {
		return c.with(Fail, "robots.txt has \"Disallow: /\" for every crawler, so nothing can be crawled.", u)
	}
	for _, s := range sitemaps {
		if strings.HasPrefix(canon(s), a.origin.String()+"/") {
			return c.with(Pass, "Crawling is allowed and the sitemap is listed ("+s+").")
		}
	}
	return c.with(Warn, "Crawling is allowed, but robots.txt has no Sitemap: line for this site.", u)
}

// parseRobots reports whether the "*" group disallows everything, and the
// Sitemap lines.
func parseRobots(body string) (blocksAll bool, sitemaps []string) {
	var agents []string
	inRules := false
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(strings.SplitN(raw, "#", 2)[0])
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)

		switch key {
		case "user-agent":
			if inRules {
				agents, inRules = nil, false
			}
			agents = append(agents, value)
		case "allow", "disallow":
			inRules = true
			if key == "disallow" && value == "/" {
				for _, agent := range agents {
					if agent == "*" {
						blocksAll = true
					}
				}
			}
		case "sitemap":
			sitemaps = append(sitemaps, value)
		}
	}
	return blocksAll, sitemaps
}

type sitemapDoc struct {
	XMLName xml.Name
	URLs    []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

const sitemapNS = "http://www.sitemaps.org/schemas/sitemap/0.9"

func (a *Auditor) readSitemap(ctx context.Context) ([]string, Check) {
	c := Check{ID: "sitemap", Group: GroupTechnical, Title: "The XML sitemap is valid"}
	root := a.origin.String() + "/sitemap.xml"

	var urls []string
	var problems []string
	seen := map[string]bool{}

	files := []string{root}
	for i := 0; i < len(files) && i < 50; i++ {
		file := files[i]
		res, err := a.get(ctx, file, sitemapMaxBytes)
		if err != nil {
			return nil, c.with(Fail, "The sitemap could not be fetched: "+err.Error(), file)
		}
		if res.status != http.StatusOK {
			return nil, c.with(Fail, fmt.Sprintf("The sitemap answered %d.", res.status), file)
		}
		if len(res.body) > sitemapMaxBytes {
			return nil, c.with(Fail, "The sitemap is over Google's 50 MB limit.", file)
		}

		var doc sitemapDoc
		if err := xml.Unmarshal(res.body, &doc); err != nil {
			return nil, c.with(Fail, "The sitemap is not valid XML: "+err.Error(), file)
		}
		if doc.XMLName.Space != sitemapNS || (doc.XMLName.Local != "urlset" && doc.XMLName.Local != "sitemapindex") {
			return nil, c.with(Fail, "The sitemap's root is not a <urlset> or <sitemapindex> in the sitemap namespace.", file)
		}

		for _, s := range doc.Sitemaps {
			if child := a.onSiteRaw(s.Loc); child != "" {
				files = append(files, child)
			} else {
				problems = append(problems, "a sitemap index entry is not on this site: "+s.Loc)
			}
		}
		for _, entry := range doc.URLs {
			loc := strings.TrimSpace(entry.Loc)
			internal := a.onSiteRaw(loc)
			switch {
			case internal == "":
				problems = append(problems, "not an absolute URL on this site: "+loc)
			case seen[canon(internal)]:
				problems = append(problems, "listed twice: "+loc)
			default:
				seen[canon(internal)] = true
				urls = append(urls, canon(internal))
			}
		}
	}

	switch {
	case len(urls) == 0:
		return nil, c.with(Fail, "The sitemap lists no pages.", root)
	case len(urls) > sitemapMaxURLs:
		return urls, c.with(Fail, fmt.Sprintf("The sitemap lists %d URLs; one file may hold 50,000.", len(urls)), root)
	case len(problems) > 0:
		return urls, c.with(Fail, fmt.Sprintf("%d problem(s) in the sitemap.", len(problems)), problems...)
	}
	return urls, c.with(Pass, fmt.Sprintf("Valid, with %d URLs, all on this site and none listed twice.", len(urls)))
}

func (a *Auditor) checkHTTPS(ctx context.Context) Check {
	c := Check{ID: "https", Group: GroupTechnical, Title: "The site is served over HTTPS"}
	if a.origin.Scheme != "https" {
		return c.with(Warn, "The site address is plain HTTP. Fine for local development, not for production.", a.origin.String())
	}

	plain := "http://" + a.origin.Host + "/"
	insecure := *a.client
	res, err := (&Auditor{origin: &url.URL{Scheme: "http", Host: a.origin.Host}, client: &insecure}).get(ctx, plain, 64<<10)
	if err != nil {
		return c.with(Pass, "HTTPS is used. (The plain-HTTP address could not be reached to check its redirect.)")
	}
	if res.status >= 300 && res.status < 400 && strings.HasPrefix(res.location, "https://"+a.origin.Host) {
		return c.with(Pass, fmt.Sprintf("HTTPS is used, and http:// answers %d to the HTTPS address.", res.status))
	}
	return c.with(Warn, fmt.Sprintf("http:// answered %d instead of redirecting to HTTPS.", res.status), plain)
}

func (c Check) with(status Status, detail string, urls ...string) Check {
	c.Status, c.Detail, c.URLs = status, detail, urls
	return c
}

// ── Checks over the crawl ───────────────────────────────────────────────────

type state struct {
	origin   *url.URL
	home     string
	pages    map[string]*Page
	order    []string
	sitemap  []string
	incoming map[string]map[string]bool
}

func (s *state) live(u string) bool {
	p := s.pages[u]
	return p != nil && p.Status == http.StatusOK && p.html
}

// indexablePages are the HTML pages a search engine may index, in crawl order.
func (s *state) indexablePages() []*Page {
	var out []*Page
	for _, u := range s.order {
		if p := s.pages[u]; p.Indexable && p.html {
			out = append(out, p)
		}
	}
	return out
}

// languageRoots are the home pages of the other languages, taken from the home
// page's hreflang alternates (e.g. /km).
func (s *state) languageRoots() []string {
	var roots []string
	if h := s.pages[s.home]; h != nil {
		for _, href := range h.Hreflang {
			if u, err := url.Parse(href); err == nil && u.Host == s.origin.Host && strings.Trim(u.Path, "/") != "" {
				roots = append(roots, "/"+strings.Trim(u.Path, "/"))
			}
		}
	}
	return roots
}

// sectionDepth is how many path segments a page has once any language prefix
// is removed: /work is 1, /km/work is 1, /work/x is 2.
func (s *state) sectionDepth(rawURL string, roots []string) int {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0
	}
	p := u.Path
	for _, root := range roots {
		if p == root || strings.HasPrefix(p, root+"/") {
			p = strings.TrimPrefix(p, root)
			break
		}
	}
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return 0
	}
	return len(strings.Split(trimmed, "/"))
}

func (s *state) checkNavigation() Check {
	c := Check{ID: "navigation", Group: GroupSitelinks, Title: "The main navigation is crawlable and leads to real pages"}
	h := s.pages[s.home]
	if h == nil || !s.live(s.home) {
		return c.with(Fail, "The home page did not load, so its navigation could not be read.", s.home)
	}
	if len(h.navTargets) == 0 {
		return c.with(Fail, "The home page has no <nav> with links in its HTML. Navigation built only by JavaScript is invisible to a crawler.")
	}

	var pages []string
	for _, t := range h.navTargets {
		if t != s.home {
			pages = append(pages, t)
		}
	}
	if len(pages) < 2 {
		return c.with(Warn, "The navigation is crawlable, but almost every link is a jump to a section of the home page (#…). "+
			"Search engines see those as the home page itself, and sitelinks are drawn from separate pages. "+
			"Linking pages such as /work or /resume from a navigation bar gives them a clearer signal.")
	}
	return c.with(Pass, fmt.Sprintf("Crawlable <nav> links to %d separate pages.", len(pages)), pages...)
}

func (s *state) checkOrphans() Check {
	c := Check{ID: "orphans", Group: GroupSitelinks, Title: "Every page in the sitemap is reachable by links"}
	var orphans []string
	for _, u := range s.sitemap {
		if p := s.pages[u]; p != nil && p.Depth < 0 {
			orphans = append(orphans, u)
		}
	}
	if len(orphans) > 0 {
		return c.with(Fail, fmt.Sprintf("%d page(s) are in the sitemap, but no link on the site leads to them.", len(orphans)), orphans...)
	}
	return c.with(Pass, "Every sitemap page can be reached by following links from the home page.")
}

func (s *state) checkClickDepth() Check {
	c := Check{ID: "click_depth", Group: GroupSitelinks, Title: "Pages are a few clicks from the home page"}
	var deep []string
	maxDepth := 0
	for _, u := range s.sitemap {
		if p := s.pages[u]; p != nil && p.Depth >= 0 {
			if p.Depth > maxDepth {
				maxDepth = p.Depth
			}
			if p.Depth > 3 {
				deep = append(deep, fmt.Sprintf("%s (%d clicks)", u, p.Depth))
			}
		}
	}
	if len(deep) > 0 {
		return c.with(Warn, "Some pages take more than 3 clicks to reach. Pages closer to the home page are treated as more important.", deep...)
	}
	return c.with(Pass, fmt.Sprintf("The furthest page is %d click(s) from the home page.", maxDepth))
}

func (s *state) checkWeakLinks() Check {
	c := Check{ID: "weak_links", Group: GroupSitelinks, Title: "Important pages have more than one link in"}
	var weak []string
	for _, u := range s.sitemap {
		p := s.pages[u]
		if p == nil || u == s.home || p.Depth < 0 || !p.Indexable {
			continue
		}
		if p.Incoming < 2 {
			weak = append(weak, fmt.Sprintf("%s (%d link in)", u, p.Incoming))
		}
	}
	if len(weak) > 0 {
		return c.with(Warn, "These pages are linked from only one other page.", weak...)
	}
	return c.with(Pass, "Every sitemap page is linked from at least two other pages.")
}

func (s *state) checkBreadcrumbs() Check {
	c := Check{ID: "breadcrumbs", Group: GroupSitelinks, Title: "Deeper pages describe where they sit (BreadcrumbList)"}
	roots := s.languageRoots()
	var deep, missing []string
	for _, p := range s.indexablePages() {
		if s.sectionDepth(p.URL, roots) < 2 {
			continue
		}
		deep = append(deep, p.URL)
		if !contains(p.StructuredData, "BreadcrumbList") {
			missing = append(missing, p.URL)
		}
	}
	switch {
	case len(deep) == 0:
		return c.with(Pass, "No page is nested deep enough to need breadcrumbs.")
	case len(missing) > 0:
		return c.with(Warn, fmt.Sprintf("%d nested page(s) have no BreadcrumbList structured data.", len(missing)), missing...)
	}
	return c.with(Pass, fmt.Sprintf("All %d nested pages carry BreadcrumbList structured data.", len(deep)))
}

func (s *state) checkSitemapPages() Check {
	c := Check{ID: "sitemap_pages", Group: GroupTechnical, Title: "The sitemap lists only live, indexable, canonical pages"}
	var bad []string
	for _, u := range s.sitemap {
		p := s.pages[u]
		switch {
		case p == nil:
			continue // not fetched: the crawl hit its page limit
		case p.Error != "":
			bad = append(bad, u+" — could not be fetched")
		case p.RedirectTo != "":
			bad = append(bad, fmt.Sprintf("%s — redirects (%d) to %s", u, p.Status, p.RedirectTo))
		case p.Status != http.StatusOK:
			bad = append(bad, fmt.Sprintf("%s — answers %d", u, p.Status))
		case !p.Indexable:
			bad = append(bad, u+" — marked noindex")
		case p.html && p.Canonical != "" && canon(p.Canonical) != u:
			bad = append(bad, u+" — canonical points elsewhere: "+p.Canonical)
		}
	}
	if len(bad) > 0 {
		return c.with(Fail, fmt.Sprintf("%d sitemap URL(s) should not be there, or should be fixed.", len(bad)), bad...)
	}
	return c.with(Pass, fmt.Sprintf("All %d sitemap URLs answer 200, are indexable and are their own canonical.", len(s.sitemap)))
}

// linksTo lists "target ← source" for every link into a page matching match.
func (s *state) linksTo(match func(*Page) bool) []string {
	var out []string
	for _, src := range s.order {
		for _, target := range s.pages[src].links {
			if p := s.pages[target]; p != nil && target != src && match(p) {
				out = append(out, fmt.Sprintf("%s ← linked from %s", target, src))
			}
		}
	}
	return out
}

func (s *state) checkBrokenLinks() Check {
	c := Check{ID: "broken_links", Group: GroupTechnical, Title: "No internal link leads to a missing page"}
	broken := s.linksTo(func(p *Page) bool { return p.Error != "" || p.Status >= 400 })
	if len(broken) > 0 {
		return c.with(Fail, fmt.Sprintf("%d link(s) lead to a page that is missing or failing.", len(broken)), broken...)
	}
	return c.with(Pass, "Every internal link leads to a page that loads.")
}

func (s *state) checkRedirectLinks() Check {
	c := Check{ID: "redirect_links", Group: GroupTechnical, Title: "Internal links point straight at the final URL"}
	redirected := s.linksTo(func(p *Page) bool { return p.RedirectTo != "" })
	if len(redirected) > 0 {
		return c.with(Warn, fmt.Sprintf("%d link(s) go through a redirect. Link to the final address instead.", len(redirected)), redirected...)
	}
	return c.with(Pass, "No internal link goes through a redirect.")
}

func (s *state) checkCanonicals() Check {
	c := Check{ID: "canonical", Group: GroupTechnical, Title: "Every page declares a valid canonical URL"}
	var missing, offsite, dead, other []string
	for _, p := range s.indexablePages() {
		if p.Canonical == "" {
			missing = append(missing, p.URL)
			continue
		}
		target := canon(p.Canonical)
		u, err := url.Parse(target)
		if err != nil || u.Scheme != s.origin.Scheme || u.Host != s.origin.Host {
			offsite = append(offsite, p.URL+" → "+p.Canonical)
			continue
		}
		if target == p.URL {
			continue
		}
		if t := s.pages[target]; t != nil && t.Status != http.StatusOK {
			dead = append(dead, fmt.Sprintf("%s → %s (answers %d)", p.URL, target, t.Status))
			continue
		}
		other = append(other, p.URL+" → "+target)
	}

	switch {
	case len(offsite)+len(dead) > 0:
		return c.with(Fail, "Some canonicals point off this site, or at a page that does not load.", append(offsite, dead...)...)
	case len(missing) > 0:
		return c.with(Warn, fmt.Sprintf("%d page(s) have no canonical tag.", len(missing)), missing...)
	case len(other) > 0:
		return c.with(Warn, "Some pages name another page as canonical, so they will not be indexed themselves. Fine for true duplicates only.", other...)
	}
	return c.with(Pass, "Every indexable page names itself as canonical.")
}

func (s *state) checkHreflang() Check {
	c := Check{ID: "hreflang", Group: GroupTechnical, Title: "Language versions link to each other (hreflang)"}
	var withAlternates int
	var dead, oneWay []string
	for _, p := range s.indexablePages() {
		if len(p.Hreflang) == 0 {
			continue
		}
		withAlternates++
		for lang, href := range p.Hreflang {
			target := canon(href)
			alt := s.pages[target]
			if alt == nil {
				continue // not crawled: off-site or past the page limit
			}
			if !s.live(target) {
				dead = append(dead, fmt.Sprintf("%s → %s %s (answers %d)", p.URL, lang, target, alt.Status))
				continue
			}
			back := false
			for _, h := range alt.Hreflang {
				if canon(h) == p.URL {
					back = true
					break
				}
			}
			if !back {
				oneWay = append(oneWay, fmt.Sprintf("%s → %s %s, which does not link back", p.URL, lang, target))
			}
		}
	}

	switch {
	case withAlternates == 0:
		return c.with(Pass, "No page declares language versions, so there is nothing to check.")
	case len(dead) > 0:
		return c.with(Fail, "Some hreflang alternates do not load.", dead...)
	case len(oneWay) > 0:
		return c.with(Warn, "hreflang only counts when both pages name each other.", unique(oneWay)...)
	}
	return c.with(Pass, fmt.Sprintf("%d page(s) declare language versions, and every pair links both ways.", withAlternates))
}

func (s *state) checkTitles() Check {
	c := Check{ID: "titles", Group: GroupOnPage, Title: "Every page has its own title"}
	return s.uniqueField(c, func(p *Page) string { return p.Title }, Fail, "title")
}

func (s *state) checkDescriptions() Check {
	c := Check{ID: "descriptions", Group: GroupOnPage, Title: "Every page has its own meta description"}
	return s.uniqueField(c, func(p *Page) string { return p.Description }, Warn, "meta description")
}

func (s *state) uniqueField(c Check, field func(*Page) string, ifMissing Status, name string) Check {
	var missing []string
	byValue := map[string][]string{}
	for _, p := range s.indexablePages() {
		v := field(p)
		if v == "" {
			missing = append(missing, p.URL)
			continue
		}
		byValue[v] = append(byValue[v], p.URL)
	}
	// Text shared only by language versions of one page is an untranslated
	// field, not two pages competing — the fix is a translation, so it is
	// reported apart from true duplicates.
	var dupes, untranslated []string
	for v, urls := range byValue {
		if len(urls) < 2 {
			continue
		}
		line := fmt.Sprintf("%q on %s", v, strings.Join(urls, ", "))
		if s.allAlternates(urls) {
			untranslated = append(untranslated, line)
		} else {
			dupes = append(dupes, line)
		}
	}
	sort.Strings(dupes)
	sort.Strings(untranslated)

	switch {
	case len(missing) > 0:
		return c.with(ifMissing, fmt.Sprintf("%d page(s) have no %s.", len(missing), name), missing...)
	case len(dupes) > 0:
		return c.with(Warn, fmt.Sprintf("Some pages share a %s. Each page should describe itself.", name), dupes...)
	case len(untranslated) > 0:
		return c.with(Warn, fmt.Sprintf("%d page(s) show the same %s in every language. Translate it, so each language version describes itself in its own language.", len(untranslated), name), untranslated...)
	}
	return c.with(Pass, fmt.Sprintf("Every indexable page has a %s of its own.", name))
}

// allAlternates reports whether every page in urls names every other one as a
// language version (hreflang).
func (s *state) allAlternates(urls []string) bool {
	for _, a := range urls {
		p := s.pages[a]
		if p == nil {
			return false
		}
		alts := map[string]bool{}
		for _, h := range p.Hreflang {
			alts[canon(h)] = true
		}
		for _, b := range urls {
			if a != b && !alts[b] {
				return false
			}
		}
	}
	return true
}

func (s *state) checkHeadings() Check {
	c := Check{ID: "h1", Group: GroupOnPage, Title: "Every page has exactly one main heading (h1)"}
	var bad []string
	for _, p := range s.indexablePages() {
		if p.H1 != 1 {
			bad = append(bad, fmt.Sprintf("%s (%d h1)", p.URL, p.H1))
		}
	}
	if len(bad) > 0 {
		return c.with(Warn, fmt.Sprintf("%d page(s) have no h1, or more than one.", len(bad)), bad...)
	}
	return c.with(Pass, "Every indexable page has one h1.")
}

func (s *state) checkStructuredData() Check {
	c := Check{ID: "structured_data", Group: GroupOnPage, Title: "Structured data (JSON-LD) is valid"}
	var invalid []string
	for _, u := range s.order {
		if p := s.pages[u]; p.invalidJSONLD > 0 {
			invalid = append(invalid, fmt.Sprintf("%s (%d invalid block(s))", u, p.invalidJSONLD))
		}
	}
	if len(invalid) > 0 {
		return c.with(Fail, "Some JSON-LD blocks are not valid JSON, so search engines ignore them.", invalid...)
	}
	if h := s.pages[s.home]; h != nil && s.live(s.home) && !contains(h.StructuredData, "WebSite") {
		return c.with(Warn, "The home page has no WebSite structured data. Google reads the site name shown in results from it.", s.home)
	}
	return c.with(Pass, "All JSON-LD is valid, and the home page names the site (WebSite).")
}

func (s *state) checkImageAlt() Check {
	c := Check{ID: "image_alt", Group: GroupOnPage, Title: "Images have alt text"}
	var bad []string
	for _, p := range s.indexablePages() {
		if p.ImagesWithoutAlt > 0 {
			bad = append(bad, fmt.Sprintf("%s (%d image(s))", p.URL, p.ImagesWithoutAlt))
		}
	}
	if len(bad) > 0 {
		return c.with(Warn, "Some images have no alt attribute. Decorative images should still carry alt=\"\".", bad...)
	}
	return c.with(Pass, "Every image has an alt attribute.")
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func unique(list []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range list {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

// ErrNotConfigured is returned when no site address is set.
var ErrNotConfigured = errors.New("seoaudit: no site URL configured")
