-- The three positioning statements under the hero (§1).
--
-- Last of the homepage content still living in the frontend's typed files.
-- Same shape as principles, minus the icon: these are numbered 01/02/03 on the
-- page, and that number is their position rather than a stored value — so
-- reordering renumbers them and nothing can disagree with itself.

CREATE TABLE IF NOT EXISTS pillars (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    slug        VARCHAR(80)  NOT NULL,
    title       VARCHAR(160) NOT NULL,
    description TEXT         NOT NULL,
    sort_order  INT          NOT NULL DEFAULT 0,

    PRIMARY KEY (id),
    UNIQUE KEY uq_pillars_slug (slug),
    KEY idx_pillars_order (sort_order)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

INSERT INTO pillars (slug, title, description, sort_order) VALUES
  ('experience', 'Experience', 'Professional full-time software development since 2020.', 0),
  ('technical-breadth', 'Technical breadth', 'Backend, frontend, mobile, databases, APIs, infrastructure, WordPress, SEO, and cloud deployment.', 1),
  ('production-ownership', 'Production ownership', 'Architecture, development, database, security, deployment, monitoring, backup, and maintenance — not just writing code.', 2)
ON DUPLICATE KEY UPDATE slug = slug;
