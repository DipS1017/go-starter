-- name: CreateUser :one
INSERT INTO
    users (
        name,
        email,
        phone_number,
        username,
        PASSWORD,
        google_id,
        apple_id,
        is_email_verified
    )
VALUES
    (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
    ) RETURNING id,
    name,
    username,
    email,
    phone_number,
    role;

-- name: GetUserByID :one
SELECT
   u.id,
   u.name,
   u.email,
   u.phone_number,
   u.username,
   u.google_id,
   u.apple_id,
   u.is_email_verified,
   u.profile_image,
   u.role,
   u.user_type,
   u.STATUS,
    (
        u.PASSWORD IS NOT NULL
        AND length(u.PASSWORD) > 0
    ) AS has_password
FROM
    users u
WHERE
    u.id = $1;

-- name: GetUserEmailByID :one
SELECT
    email
FROM
    users
WHERE
    id = $1;

-- name: CreateResetToken :one
UPDATE
    users
SET
    reset_token = $1,
    reset_expires = $2
WHERE
    email = $3 RETURNING id,
    name,
    email,
    reset_token,
    role,
    reset_expires;

-- name: GetUserByResetToken :one
SELECT
    id,
    reset_expires
FROM
    users
WHERE
    reset_token = $1;

-- name: UpdatePasswordReset :one
UPDATE
    users
SET
    PASSWORD = $1,
    reset_token = NULL,
    reset_expires = NULL
WHERE
    (
        id = $2
        AND reset_token = $3
    ) RETURNING id,
    name;

-- name: GetUserByEmail :one
SELECT
    id,
    name,
    email,
    phone_number,
    username,
    PASSWORD,
    google_id,
    apple_id,
    role,
    STATUS,
    is_email_verified
FROM
    users
WHERE
    email = $1;


-- name: GetUserByGoogleID :one
SELECT
    id,
    name,
    email,
    phone_number,
    username,
    PASSWORD,
    google_id,
    apple_id,
    is_email_verified
FROM
    users
WHERE
    google_id = $1;

-- name: UpdateEmailStatus :one
UPDATE
    users
SET
    is_email_verified = $1
WHERE
    id = $2 RETURNING id,
    is_email_verified,
    email;

-- name: GetUserByAppleID :one
SELECT
    id,
    name,
    email,
    phone_number,
    username,
    PASSWORD,
    google_id,
    apple_id,
    is_email_verified
FROM
    users
WHERE
    apple_id = $1;

-- name: UpdateProfileImage :one
UPDATE
    users
SET
    profile_image = $1
WHERE
    id = $2 RETURNING id,
    name,
    email,
    phone_number,
    username,
    profile_image;

-- name: CreateUserSession :one
WITH existing_session AS (
    SELECT
        EXISTS (
            SELECT
                1
            FROM
                user_sessions
            WHERE
                user_sessions.user_id = $1
        ) AS has_sessions
),
new_session AS (
    INSERT INTO
        user_sessions (
            user_id,
            refresh_token_hash,
            user_agent,
            ip_address,
            expires_at
        )
    VALUES
        ($1, $2, $3, $4, $5) RETURNING id
)
SELECT
    new_session.id,
    NOT existing_session.has_sessions AS first_login
FROM
    new_session,
    existing_session;

-- name: GetSessionByID :one
SELECT
    id,
    user_id,
    revoked_at,
    last_seen_at,
    expires_at
FROM
    user_sessions
WHERE
    id = $1
    AND expires_at > NOW();

-- name: UpdateSessionLastSeen :exec
UPDATE
    user_sessions
SET
    last_seen_at = NOW()
WHERE
    id = $1;

-- name: GetSessionByRefreshTokenHash :one
SELECT
    id,
    user_id
FROM
    user_sessions
WHERE
    refresh_token_hash = $1
    AND revoked_at IS NULL
    AND expires_at > NOW();

-- name: GetOfflineUsers :many
SELECT DISTINCT user_id
FROM user_sessions
WHERE last_seen_at < NOW() - INTERVAL '1 minute'
LIMIT 1000;

-- name: RevokeSessionByID :one
UPDATE
    user_sessions
SET
    revoked_at = NOW()
WHERE
    id = $1 RETURNING user_id;

-- name: DeleteOldSessions :exec
DELETE FROM
    user_sessions
WHERE
    (
        revoked_at IS NOT NULL
        AND revoked_at < NOW() - INTERVAL '90 days'
    )
    OR (expires_at < NOW() - INTERVAL '90 days');

-- name: UpdateUser :exec
UPDATE
    users
SET
    name = COALESCE(sqlc.narg(name), name),
    username = COALESCE(sqlc.narg(username), username),
    phone_number = COALESCE(sqlc.narg(phone_number), phone_number),
    stripe_customer_id = COALESCE(sqlc.narg(stripe_customer_id), stripe_customer_id),
    updated_at = NOW()
WHERE
    id = sqlc.arg(id);

-- name: GetPasswordByID :one
SELECT
    PASSWORD
FROM
    users
WHERE
    id = $1;

-- name: UpdatePasswordByID :exec
UPDATE
    users
SET
    PASSWORD = $1
WHERE
    id = $2;

-- name: GetUserCredentialByID :one
SELECT
    email,
    PASSWORD,
    name
FROM
    users
WHERE
    id = $1;

-- name: ChangeAccountStatus :one
UPDATE
    users
SET
    STATUS = $1
WHERE
    id = $2 RETURNING id,
    name,
    email;


-- name: GetAllUserIDs :many
SELECT
    id
FROM
    users
WHERE
    id <> @user_id;

-- name: GetUserStripeInfoByID :one
SELECT
    email,
    name,
    stripe_customer_id
FROM
    users
WHERE
    id = $1;

-- name: GetUserStripeInfoByEmail :one
SELECT
    id,
    stripe_customer_id
FROM
    users
WHERE
    email = $1;
