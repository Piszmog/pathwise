-- name: ExistsHNStory :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      hn_stories
    WHERE
      id = ?
  );

-- name: InsertHNStory :exec
INSERT INTO
  hn_stories (posted_at, title, id)
VALUES
  (?, ?, ?);

-- name: ExistsHNComment :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      hn_comments
    WHERE
      id = ?
  );

-- name: InsertHNComment :exec
INSERT INTO
  hn_comments (commented_at, value, id, hn_story_id)
VALUES
  (?, ?, ?, ?);
