ALTER TABLE users
DROP COLUMN initial_ip_address;

ALTER TABLE sessions
DROP COLUMN ip_address;
