-- "Built for Production" (§19) moves out of the frontend's typed files and into
-- the database, like every other section, so it can be edited and translated at
-- runtime instead of needing a deploy.
--
-- Same shape as services: a slug for a stable key, an Iconify name, and an
-- explicit order. The section is deliberately small, so there is no pagination
-- pressure and no need for anything more.

CREATE TABLE IF NOT EXISTS principles (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    slug        VARCHAR(80)  NOT NULL,
    title       VARCHAR(160) NOT NULL,
    description TEXT         NOT NULL,
    icon        VARCHAR(80)  NOT NULL,
    sort_order  INT          NOT NULL DEFAULT 0,

    PRIMARY KEY (id),
    UNIQUE KEY uq_principles_slug (slug),
    KEY idx_principles_order (sort_order)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- The four that were hard-coded, so the section does not empty out on upgrade.
SET @now = UTC_TIMESTAMP();
INSERT INTO principles (slug, title, description, icon, sort_order) VALUES
  ('architecture', 'Architecture', 'Designed before implementation, so the structure survives the second and third feature.', 'lucide:compass', 0),
  ('security', 'Security', 'Considered from the application layer down to the infrastructure it runs on.', 'lucide:lock', 1),
  ('reliability', 'Reliability', 'Errors, backups, monitoring, and recovery are part of the system, not an afterthought.', 'lucide:heart-pulse', 2),
  ('maintainability', 'Maintainability', 'Clean structure and documentation so future development does not start with archaeology.', 'lucide:wrench', 3)
ON DUPLICATE KEY UPDATE slug = slug;
