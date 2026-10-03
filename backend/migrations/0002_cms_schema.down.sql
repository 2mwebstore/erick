ALTER TABLE contact_messages
    DROP KEY idx_contact_messages_unread,
    DROP COLUMN archived_at,
    DROP COLUMN read_at;

DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS site_settings;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS capabilities;
DROP TABLE IF EXISTS experience_entries;
DROP TABLE IF EXISTS case_studies;
DROP TABLE IF EXISTS project_technologies;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
