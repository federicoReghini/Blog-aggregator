-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeed :one
SELECT * FROM feeds
WHERE feeds.id = $1;

-- name: GetFeeds :many
SELECT *, users.name as user_name FROM feeds
INNER JOIN users
ON feeds.user_id = users.id
ORDER By feeds.created_at DESC;

-- name: GetFeedByUrl :one
SELECT * FROM feeds
WHERE feeds.url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds 
SET updated_at = now(), last_fetched_at = now()
WHERE feeds.id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY feeds.last_fetched_at ASC NULLS FIRST
LIMIT 1;

