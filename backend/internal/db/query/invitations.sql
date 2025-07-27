-- name: CreateInvitation :one
INSERT INTO invitations (
    email, username, token, invited_by, permissions, create_media_user, expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetInvitationByToken :one
SELECT * FROM invitations WHERE token = ? AND status = 'pending' AND expires_at > datetime('now');

-- name: GetInvitationByEmail :one
SELECT * FROM invitations WHERE email = ? ORDER BY created_at DESC LIMIT 1;

-- name: GetInvitationByID :one
SELECT * FROM invitations WHERE id = ?;

-- name: GetInvitationsByInviter :many
SELECT * FROM invitations WHERE invited_by = ? ORDER BY created_at DESC;

-- name: GetAllInvitations :many
SELECT 
    i.*,
    u.username as inviter_username
FROM invitations i
LEFT JOIN users u ON i.invited_by = u.id
ORDER BY i.created_at DESC;

-- name: GetPendingInvitations :many
SELECT * FROM invitations 
WHERE status = 'pending' AND expires_at > datetime('now')
ORDER BY created_at DESC;

-- name: UpdateInvitationStatus :one
UPDATE invitations 
SET status = ?, updated_at = CURRENT_TIMESTAMP, accepted_at = CASE WHEN ? = 'accepted' THEN CURRENT_TIMESTAMP ELSE accepted_at END
WHERE token = ? RETURNING *;

-- name: CancelInvitation :one
UPDATE invitations 
SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
WHERE id = ? RETURNING *;

-- name: ExpireOldInvitations :exec
UPDATE invitations 
SET status = 'expired', updated_at = CURRENT_TIMESTAMP
WHERE status = 'pending' AND expires_at <= datetime('now');

-- name: DeleteInvitation :exec
DELETE FROM invitations WHERE id = ?;

-- name: GetInvitationStats :one
SELECT 
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_count,
    COUNT(CASE WHEN status = 'accepted' THEN 1 END) as accepted_count,
    COUNT(CASE WHEN status = 'expired' THEN 1 END) as expired_count,
    COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_count,
    COUNT(*) as total_count
FROM invitations;