-- name: CreatePost :one
INSERT INTO posts(id, created_at, updated_at, published_at, title, url, description, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT * FROM posts
ORDER BY posts.published_at 
LIMIT $1;
