-- Fields the redesigned Contact and Work Experience sections need.
--
-- Contact gains a phone number and a subject line. Both are nullable: messages
-- already in the inbox predate them, and a null there means "not asked for",
-- which is true, rather than an empty string pretending it was left blank.
ALTER TABLE contact_messages
    ADD COLUMN phone   VARCHAR(40)  NULL AFTER email,
    ADD COLUMN subject VARCHAR(160) NULL AFTER phone;

-- Work Experience becomes an employment history rather than a list of
-- capabilities, so each entry needs an employer, a period, a place, and the
-- stack it was built on.
--
-- `technologies` is JSON, matching how capabilities store their tool list. A
-- join table would buy per-technology querying that nothing asks for.
ALTER TABLE experience_entries
    ADD COLUMN company      VARCHAR(160) NULL AFTER title,
    ADD COLUMN period       VARCHAR(40)  NULL AFTER company,
    ADD COLUMN location     VARCHAR(120) NULL AFTER period,
    -- Drives the "Current" badge. A flag rather than parsing the period text,
    -- which would break the moment someone writes "Present" differently.
    ADD COLUMN is_current   TINYINT(1)   NOT NULL DEFAULT 0 AFTER location,
    ADD COLUMN technologies JSON         NULL AFTER is_current;
