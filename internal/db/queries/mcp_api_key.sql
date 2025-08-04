-- name: InsertMcpAPIKey :one
INSERT INTO mcp_api_keys (user_id, key_hash)
VALUES (?, ?)
RETURNING id, created_at;

-- name: GetMcpAPIKeyByHash :one
SELECT id, user_id, key_hash, created_at, updated_at 
FROM mcp_api_keys 
WHERE key_hash = ?;

-- name: DeleteMcpAPIKeyByUserID :exec
DELETE FROM mcp_api_keys WHERE user_id = ?;