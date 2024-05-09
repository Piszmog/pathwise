ALTER TABLE users
ADD COLUMN initial_ip_address TEXT NOT NULL DEFAULT '';

UPDATE users
SET initial_ip_address = '';

ALTER TABLE sessions
ADD COLUMN ip_address TEXT NOT NULL DEFAULT '';

UPDATE sessions
SET ip_address = '';
