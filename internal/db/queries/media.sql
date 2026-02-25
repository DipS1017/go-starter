
-- name: CreateMedia :one
INSERT INTO media (name, url, type)
VALUES ($1, $2, $3)
RETURNING id,url;
-- name: DeleteMediaByID :one
DELETE FROM media
WHERE id = $1
RETURNING id;

-- name: GetUnusedMedia :many
SELECT id, url
FROM media
WHERE reference_count <= 0
  AND created_at < NOW() - INTERVAL '1 day';

-- name: DeleteMediaByIDBulk :exec
DELETE FROM media
WHERE id = ANY($1::uuid[]);

-- name: GetMediaByID :one
SELECT * FROM media
WHERE id = $1;


