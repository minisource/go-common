-- name: CreateModel :one
INSERT INTO models (field1, field2, created_by, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateModel :one
UPDATE models
SET field1 = $1, field2 = $2, modified_by = $3, modified_at = $4
WHERE id = $5 AND deleted_by IS NULL
RETURNING *;

-- name: DeleteModel :exec
UPDATE models
SET deleted_by = $1, deleted_at = $2
WHERE id = $3 AND deleted_by IS NULL;

-- name: GetModel :one
SELECT * FROM models
WHERE id = $1 AND deleted_by IS NULL;

-- name: ListModels :many
SELECT * FROM models
WHERE deleted_by IS NULL AND $1
ORDER BY $2
LIMIT $3 OFFSET $4;

-- name: CountModels :one
SELECT COUNT(*) FROM models
WHERE deleted_by IS NULL AND $1;