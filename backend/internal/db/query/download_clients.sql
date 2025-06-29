-- name: GetDownloadClients :many
SELECT id, type, name, host, port, username, password, api_key, use_ssl, created_at
FROM download_clients;

-- name: UpsertDownloadClient :exec
INSERT INTO download_clients (id, type, name, host, port, username, password, api_key, use_ssl)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    type = excluded.type,
    name = excluded.name,
    host = excluded.host,
    port = excluded.port,
    username = excluded.username,
    password = excluded.password,
    api_key = excluded.api_key,
    use_ssl = excluded.use_ssl;