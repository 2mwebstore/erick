-- Khmer (and any future language) as an overlay on the English content.
--
-- One generic table rather than a `_km` column beside every text column. The
-- case study alone has eleven translatable fields, and duplicating those across
-- projects, capabilities, services, experience and settings would mean a schema
-- change for every new language. Here a language is rows, not columns.
--
-- Nothing is required. A field with no row for a locale simply keeps its
-- English text, so a half-translated site renders completely rather than
-- showing blanks — which is the only sane default when one person is writing
-- both languages by hand.

CREATE TABLE IF NOT EXISTS content_translations (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    -- 'project', 'case_study', 'experience', 'capability', 'service', 'setting'.
    entity     VARCHAR(32)     NOT NULL,
    -- The row being translated. Settings are keyed by name, not id, so they use 0.
    entity_id  BIGINT UNSIGNED NOT NULL DEFAULT 0,
    -- The field within that row, e.g. 'title'. For a setting, the setting key.
    field      VARCHAR(64)     NOT NULL,
    locale     VARCHAR(8)      NOT NULL,
    -- Plain text, or a JSON array for list fields such as a capability's tools.
    value      TEXT            NOT NULL,
    updated_at DATETIME        NOT NULL,

    PRIMARY KEY (id),
    -- One translation per field per language; saves are upserts against this.
    UNIQUE KEY uq_content_translation (entity, entity_id, field, locale),
    -- The public read is "everything for this locale", which this serves.
    KEY idx_content_translation_locale (locale, entity)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
