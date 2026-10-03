ALTER TABLE contact_messages
    DROP COLUMN phone,
    DROP COLUMN subject;

ALTER TABLE experience_entries
    DROP COLUMN company,
    DROP COLUMN period,
    DROP COLUMN location,
    DROP COLUMN is_current,
    DROP COLUMN technologies;
