-- Contact form submissions (§29).
--
-- Only the tables that are actually needed exist: there is no users table
-- because there is no admin panel, and no projects table because project
-- content is versioned in the frontend repository instead.
CREATE TABLE IF NOT EXISTS contact_messages (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name          VARCHAR(100)    NOT NULL,
    email         VARCHAR(254)    NOT NULL,
    project_type  VARCHAR(50)     NOT NULL,
    message       TEXT            NOT NULL,
    -- Request metadata, kept so abusive traffic can be investigated after the
    -- fact. Nullable because a direct request may carry neither.
    ip_address    VARCHAR(45)     NULL,
    user_agent    VARCHAR(512)    NULL,
    created_at    DATETIME        NOT NULL,
    PRIMARY KEY (id),
    -- Supports the newest-first listing and the abuse-rate query.
    KEY idx_contact_messages_created_at (created_at),
    KEY idx_contact_messages_email (email)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
