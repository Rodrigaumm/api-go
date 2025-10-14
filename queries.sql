-- name: GetUser :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByName :one
SELECT * FROM users WHERE name = $1 LIMIT 1;

-- name: GetUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (name, password) VALUES ($1, $2) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET name = $1, password = $2, updated_at = NOW() WHERE id = $3 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- ============================================
-- Process Snapshots Queries
-- ============================================

-- name: CreateProcessSnapshot :one
INSERT INTO process_snapshots (
    user_id,
    webhook_url,
    snapshot_type,
    process_count,
    success,
    error_message
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetProcessSnapshot :one
SELECT * FROM process_snapshots WHERE id = $1 LIMIT 1;

-- name: GetProcessSnapshotsByUser :many
SELECT * FROM process_snapshots 
WHERE user_id = $1 OR user_id IS NULL
ORDER BY created_at DESC;

-- name: GetProcessSnapshotsByType :many
SELECT * FROM process_snapshots 
WHERE (user_id = $1 OR user_id IS NULL) AND snapshot_type = $2
ORDER BY created_at DESC;

-- name: UpdateProcessSnapshotCount :exec
UPDATE process_snapshots 
SET process_count = $2, updated_at = NOW() 
WHERE id = $1;

-- name: DeleteProcessSnapshot :exec
DELETE FROM process_snapshots WHERE id = $1;

-- ============================================
-- Process Info Queries
-- ============================================

-- name: CreateProcessInfo :one
INSERT INTO process_info (
    snapshot_id,
    user_id,
    process_id,
    parent_process_id,
    process_name,
    thread_count,
    handle_count,
    base_priority,
    create_time,
    user_time,
    kernel_time,
    working_set_size,
    peak_working_set_size,
    virtual_size,
    peak_virtual_size,
    read_operation_count,
    write_operation_count,
    other_operation_count,
    read_transfer_count,
    write_transfer_count,
    other_transfer_count,
    current_process_address,
    next_process_eprocess_address,
    next_process_name,
    next_process_id,
    previous_process_eprocess_address,
    previous_process_name,
    previous_process_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28
) RETURNING *;

-- name: GetProcessInfo :one
SELECT * FROM process_info WHERE id = $1 LIMIT 1;

-- name: GetProcessInfosByUser :many
SELECT * FROM process_info 
WHERE user_id = $1 OR user_id IS NULL
ORDER BY created_at DESC;

-- name: GetProcessInfosBySnapshot :many
SELECT * FROM process_info 
WHERE snapshot_id = $1
ORDER BY process_id ASC;

-- name: GetProcessInfosByProcessID :many
SELECT * FROM process_info 
WHERE (user_id = $1 OR user_id IS NULL) AND process_id = $2
ORDER BY created_at DESC;

-- name: GetProcessInfoBySnapshotAndPID :one
SELECT * FROM process_info 
WHERE snapshot_id = $1 AND process_id = $2
LIMIT 1;

-- name: DeleteProcessInfo :exec
DELETE FROM process_info WHERE id = $1;

-- ============================================
-- Process Queries (Query by PID history)
-- ============================================

-- name: CreateProcessQuery :one
INSERT INTO process_queries (
    snapshot_id,
    user_id,
    webhook_url,
    requested_pid,
    process_info_id,
    success,
    error_message
) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetProcessQuery :one
SELECT * FROM process_queries WHERE id = $1 LIMIT 1;

-- name: GetProcessQueriesByUser :many
SELECT * FROM process_queries 
WHERE user_id = $1 OR user_id IS NULL
ORDER BY created_at DESC;

-- name: GetProcessQueriesBySnapshot :many
SELECT * FROM process_queries 
WHERE snapshot_id = $1
ORDER BY created_at DESC;

-- name: GetProcessQueriesByPID :many
SELECT * FROM process_queries 
WHERE (user_id = $1 OR user_id IS NULL) AND requested_pid = $2
ORDER BY created_at DESC;

-- ============================================
-- Statistics and Analytics
-- ============================================

-- name: CountUserProcesses :one
SELECT COUNT(*) FROM process_info WHERE user_id = $1;

-- name: CountUserSnapshots :one
SELECT COUNT(*) FROM process_snapshots WHERE user_id = $1;

-- name: CountUserQueries :one
SELECT COUNT(*) FROM process_queries WHERE user_id = $1;

-- name: GetMostQueriedProcesses :many
SELECT 
    requested_pid,
    COUNT(*) as query_count
FROM process_queries
WHERE user_id = $1 OR user_id IS NULL
GROUP BY requested_pid
ORDER BY query_count DESC
LIMIT $2;

-- name: GetSnapshotStatistics :one
SELECT 
    COUNT(DISTINCT snapshot_id) as total_snapshots,
    COUNT(*) as total_processes,
    AVG(process_count) as avg_processes_per_snapshot
FROM process_info
WHERE user_id = $1 OR user_id IS NULL;
