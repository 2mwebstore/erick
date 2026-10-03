-- Removes the seeded content. Destructive: anything edited in /admin since
-- seeding is deleted too. Take a backup first (scripts/backup-db.sh).
DELETE FROM site_settings;
DELETE FROM services;
DELETE FROM capabilities;
DELETE FROM experience_entries;
-- project_technologies and case_studies cascade from projects.
DELETE FROM projects;
