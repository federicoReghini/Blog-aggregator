-- +goose Up
CREATE TABLE posts (
id UUID PRIMARY KEY,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP NOT NULL,
published_at TIMESTAMP NOT NULL,
title TEXT,
url TEXT UNIQUE,
description TEXT,
feed_id UUID REFERENCES feeds(id)  ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
