-- name: CreateUser :one
INSERT INTO users (
    user_id,
    email,
    firstName,
    lastName,
    timezone,
    created_at,
    updated_at
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    NOW(),
    NOW()
) RETURNING *; 
