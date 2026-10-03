package models

import "time"

// Project mirrors the shape the frontend renders, so the API response needs no
// translation layer in the client.
type Project struct {
	ID           int64      `json:"id"`
	Slug         string     `json:"slug"`
	Title        string     `json:"title"`
	Category     string     `json:"category"`
	Description  string     `json:"description"`
	Summary      string     `json:"summary,omitempty"`
	Image        string     `json:"image,omitempty"`
	ImageAlt     string     `json:"imageAlt,omitempty"`
	Year         string     `json:"year,omitempty"`
	LiveURL      string     `json:"liveUrl,omitempty"`
	GithubURL    string     `json:"githubUrl,omitempty"`
	Featured     bool       `json:"featured"`
	Published    bool       `json:"published"`
	Order        int        `json:"order"`
	Technologies []string   `json:"technologies"`
	CaseStudy    *CaseStudy `json:"caseStudy,omitempty"`
	UpdatedAt    time.Time  `json:"updatedAt"`

	// Translations is locale → field → text, carried only on the admin payload
	// and on saves. The public payload has them already applied, so it omits
	// this entirely rather than shipping every language to every visitor.
	Translations map[string]map[string]string `json:"translations,omitempty"`
}

type TechnicalDecision struct {
	Decision  string `json:"decision"`
	Rationale string `json:"rationale"`
}

type CaseStudy struct {
	Overview     string              `json:"overview,omitempty"`
	Context      string              `json:"context,omitempty"`
	Problem      string              `json:"problem,omitempty"`
	Role         string              `json:"role,omitempty"`
	Approach     string              `json:"approach,omitempty"`
	Architecture string              `json:"architecture,omitempty"`
	Database     string              `json:"database,omitempty"`
	API          string              `json:"api,omitempty"`
	Security     string              `json:"security,omitempty"`
	Deployment   string              `json:"deployment,omitempty"`
	Outcome      string              `json:"outcome,omitempty"`
	Features     []string            `json:"features,omitempty"`
	Challenges   []string            `json:"challenges,omitempty"`
	Decisions    []TechnicalDecision `json:"decisions,omitempty"`
}

type ExperienceEntry struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
	Title string `json:"title"`
	// Employer, period, and place. All optional: an entry can still be a stage
	// of a career rather than a job, which is what these used to be.
	Company      string   `json:"company,omitempty"`
	Period       string   `json:"period,omitempty"`
	Location     string   `json:"location,omitempty"`
	Current      bool     `json:"current"`
	Technologies []string `json:"technologies"`
	Description  string   `json:"description,omitempty"`
	Order        int      `json:"order"`

	// Translations is locale → field → text, carried only on the admin payload
	// and on saves. The public payload has them already applied, so it omits
	// this entirely rather than shipping every language to every visitor.
	Translations map[string]map[string]string `json:"translations,omitempty"`
}

type Capability struct {
	ID          int64    `json:"id"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Items       []string `json:"items"`
	Order       int      `json:"order"`

	// Translations is locale → field → text, carried only on the admin payload
	// and on saves. The public payload has them already applied, so it omits
	// this entirely rather than shipping every language to every visitor.
	Translations map[string]map[string]string `json:"translations,omitempty"`
}

type Service struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`

	// Translations is locale → field → text, carried only on the admin payload
	// and on saves. The public payload has them already applied, so it omits
	// this entirely rather than shipping every language to every visitor.
	Translations map[string]map[string]string `json:"translations,omitempty"`
}

// Principle is one card in "Built for Production" (§19).
type Principle struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`

	// Translations is locale → field → text, carried only on the admin payload
	// and on saves. The public payload has them already applied, so it omits
	// this entirely rather than shipping every language to every visitor.
	Translations map[string]map[string]string `json:"translations,omitempty"`
}

// Pillar is one of the numbered positioning statements under the hero (§1).
//
// The 01/02/03 label is its position on the page, not a stored field, so
// reordering renumbers them and nothing can disagree with itself.
type Pillar struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`

	Translations map[string]map[string]string `json:"translations,omitempty"`
}

// SiteContent is the single payload the public site reads. One request rather
// than five keeps the render path simple and makes the whole thing cacheable as
// a unit.
type SiteContent struct {
	Settings     map[string]string `json:"settings"`
	Projects     []Project         `json:"projects"`
	Experience   []ExperienceEntry `json:"experience"`
	Capabilities []Capability      `json:"capabilities"`
	Services     []Service         `json:"services"`
	Principles   []Principle       `json:"principles"`
	Pillars      []Pillar          `json:"pillars"`
	GeneratedAt  time.Time         `json:"generatedAt"`

	// Locale this payload has been rendered in. The public site caches per
	// locale, so it has to be able to tell which one it is holding.
	Locale string `json:"locale"`

	// Settings have no row id, so their translations cannot hang off a model
	// the way the others do. Admin payload only.
	SettingTranslations map[string]map[string]string `json:"settingTranslations,omitempty"`
}
