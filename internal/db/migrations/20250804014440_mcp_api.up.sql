CREATE TABLE IF NOT EXISTS mcp_api_keys (
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	key_hash TEXT NOT NULL UNIQUE,
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL UNIQUE,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS mcp_api_keys_user_id_idx ON mcp_api_keys(user_id);
CREATE INDEX IF NOT EXISTS mcp_api_keys_key_hash_idx ON mcp_api_keys(key_hash);
