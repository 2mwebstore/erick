-- Content management schema (§28, §29).
--
-- This moves site content out of the frontend's typed files and into MySQL so it
-- can be edited at runtime. The trade-off is deliberate and documented in
-- docs/ARCHITECTURE.md: the database is now on the read path for every page, so
-- the frontend caches aggressively and falls back to its bundled defaults if the
-- API is unreachable.

-- ── Identity ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS users (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email         VARCHAR(254)    NOT NULL,
    name          VARCHAR(100)    NOT NULL,
    -- bcrypt output, always 60 bytes. Never a plaintext or reversible value.
    password_hash CHAR(60)        NOT NULL,
    role          ENUM('admin','editor') NOT NULL DEFAULT 'editor',
    is_active     TINYINT(1)      NOT NULL DEFAULT 1,
    last_login_at DATETIME        NULL,
    created_at    DATETIME        NOT NULL,
    updated_at    DATETIME        NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Sessions are server-side so a logout, a password change or a compromise can
-- revoke access immediately — a stateless token cannot be withdrawn.
CREATE TABLE IF NOT EXISTS sessions (
    -- SHA-256 of the session token. The token itself only ever exists in the
    -- cookie, so a database leak does not hand over live sessions.
    token_hash   CHAR(64)        NOT NULL,
    user_id      BIGINT UNSIGNED NOT NULL,
    expires_at   DATETIME        NOT NULL,
    created_at   DATETIME        NOT NULL,
    last_seen_at DATETIME        NOT NULL,
    ip_address   VARCHAR(45)     NULL,
    user_agent   VARCHAR(512)    NULL,
    PRIMARY KEY (token_hash),
    KEY idx_sessions_user (user_id),
    KEY idx_sessions_expiry (expires_at),
    CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ── Content ─────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS projects (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    slug        VARCHAR(80)     NOT NULL,
    title       VARCHAR(120)    NOT NULL,
    category    VARCHAR(80)     NOT NULL,
    description TEXT            NULL,
    summary     TEXT            NULL,
    image       VARCHAR(255)    NULL,
    image_alt   VARCHAR(255)    NULL,
    year        VARCHAR(20)     NULL,
    live_url    VARCHAR(500)    NULL,
    github_url  VARCHAR(500)    NULL,
    featured    TINYINT(1)      NOT NULL DEFAULT 0,
    -- Unpublished projects are hidden from the public API but stay editable.
    published   TINYINT(1)      NOT NULL DEFAULT 1,
    sort_order  INT             NOT NULL DEFAULT 99,
    created_at  DATETIME        NOT NULL,
    updated_at  DATETIME        NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_projects_slug (slug),
    KEY idx_projects_listing (published, featured, sort_order)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS project_technologies (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    project_id BIGINT UNSIGNED NOT NULL,
    name       VARCHAR(60)     NOT NULL,
    sort_order INT             NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uq_project_technology (project_id, name),
    KEY idx_project_technologies_order (project_id, sort_order),
    CONSTRAINT fk_project_technologies_project
        FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- One row per project. Repeating groups (features, challenges, decisions) are
-- JSON rather than child tables: they are only ever read and written as a whole
-- list, never queried across projects, so three more tables would buy nothing.
CREATE TABLE IF NOT EXISTS case_studies (
    project_id     BIGINT UNSIGNED NOT NULL,
    overview       TEXT NULL,
    context        TEXT NULL,
    problem        TEXT NULL,
    role           TEXT NULL,
    approach       TEXT NULL,
    architecture   TEXT NULL,
    database_notes TEXT NULL,
    api_notes      TEXT NULL,
    security       TEXT NULL,
    deployment     TEXT NULL,
    outcome        TEXT NULL,
    features       JSON NULL,
    challenges     JSON NULL,
    decisions      JSON NULL,
    updated_at     DATETIME NOT NULL,
    PRIMARY KEY (project_id),
    CONSTRAINT fk_case_studies_project
        FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS experience_entries (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    label       VARCHAR(40)  NOT NULL,
    title       VARCHAR(160) NOT NULL,
    description TEXT         NULL,
    sort_order  INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    KEY idx_experience_order (sort_order)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS capabilities (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    slug        VARCHAR(60)  NOT NULL,
    title       VARCHAR(120) NOT NULL,
    description TEXT         NOT NULL,
    icon        VARCHAR(80)  NOT NULL,
    items       JSON         NOT NULL,
    sort_order  INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uq_capabilities_slug (slug),
    KEY idx_capabilities_order (sort_order)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS services (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    slug        VARCHAR(60)  NOT NULL,
    title       VARCHAR(120) NOT NULL,
    description TEXT         NOT NULL,
    icon        VARCHAR(80)  NOT NULL,
    sort_order  INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uq_services_slug (slug),
    KEY idx_services_order (sort_order)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Key/value so a new setting needs no migration.
CREATE TABLE IF NOT EXISTS site_settings (
    setting_key   VARCHAR(64) NOT NULL,
    setting_value TEXT        NULL,
    updated_at    DATETIME    NOT NULL,
    PRIMARY KEY (setting_key)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ── Operations ──────────────────────────────────────────────────────────────

-- actor_email is denormalised on purpose: the log must still say who did what
-- after that user has been deleted.
CREATE TABLE IF NOT EXISTS audit_log (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id     BIGINT UNSIGNED NULL,
    actor_email VARCHAR(254) NOT NULL,
    action      VARCHAR(40)  NOT NULL,
    entity      VARCHAR(40)  NOT NULL,
    entity_id   VARCHAR(80)  NULL,
    detail      JSON         NULL,
    ip_address  VARCHAR(45)  NULL,
    created_at  DATETIME     NOT NULL,
    PRIMARY KEY (id),
    KEY idx_audit_created (created_at),
    KEY idx_audit_entity (entity, entity_id),
    CONSTRAINT fk_audit_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- The inbox needs read and archive state now that it is managed from /admin.
ALTER TABLE contact_messages
    ADD COLUMN read_at     DATETIME NULL AFTER created_at,
    ADD COLUMN archived_at DATETIME NULL AFTER read_at,
    ADD KEY idx_contact_messages_unread (read_at, created_at);
