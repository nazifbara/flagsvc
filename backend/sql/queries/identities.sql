-- name: CreateAuth :one
INSERT INTO identities (id, email, passwordHash, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    NOW(),
    NOW()
) RETURNING *; 
