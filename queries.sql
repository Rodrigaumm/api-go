-- name: GetUsers :many
SELECT id, name, password, created_at, updated_at FROM users
ORDER BY id;

-- name: GetUser :one
SELECT id, name, password, created_at, updated_at FROM users
WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
  name, password
) VALUES (
  $1, $2
)
RETURNING id, name, password, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
  set name = $2,
  password = $3,
  updated_at = NOW()
WHERE id = $1
RETURNING id, name, password, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetUserByName :one
SELECT id, name, password, created_at, updated_at FROM users
WHERE name = $1 LIMIT 1;

-- ============================================================================
-- Process Info Queries
-- ============================================================================

-- name: CreateProcessInfo :one
INSERT INTO process_info (
    user_id, process_id, parent_process_id, process_name, thread_count, handle_count,
    base_priority, create_time, user_time, kernel_time, working_set_size, peak_working_set_size,
    virtual_size, peak_virtual_size, pagefile_usage, peak_pagefile_usage, page_fault_count,
    read_operation_count, write_operation_count, other_operation_count, read_transfer_count,
    write_transfer_count, other_transfer_count, current_process_address,
    next_process_eprocess_address, next_process_name, next_process_id,
    previous_process_eprocess_address, previous_process_name, previous_process_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30
)
RETURNING *;

-- name: GetProcessInfosByUser :many
SELECT * FROM process_info
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetProcessInfo :one
SELECT * FROM process_info
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: GetProcessInfoByID :one
SELECT * FROM process_info
WHERE id = $1
LIMIT 1;

-- name: DeleteProcessInfo :exec
DELETE FROM process_info
WHERE id = $1 AND user_id = $2;

-- name: GetProcessInfosByProcessID :many
SELECT * FROM process_info
WHERE user_id = $1 AND process_id = $2
ORDER BY created_at DESC;

-- name: GetAllProcessInfos :many
SELECT * FROM process_info
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetProcessInfosByUserPaginated :many
SELECT * FROM process_info
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- ============================================================================
-- Process Iteration History Queries
-- ============================================================================

-- name: CreateProcessIterationHistory :one
INSERT INTO process_iteration_history (
    user_id, webhook_url, process_count, success, error_message
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetProcessIterationHistoryByUser :many
SELECT * FROM process_iteration_history
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetProcessIterationHistory :one
SELECT * FROM process_iteration_history
WHERE id = $1
LIMIT 1;

-- name: GetAllProcessIterationHistory :many
SELECT * FROM process_iteration_history
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetProcessIterationHistoryByUserPaginated :many
SELECT * FROM process_iteration_history
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- ============================================================================
-- Iteration Processes (Link Table) Queries
-- ============================================================================

-- name: CreateIterationProcess :one
INSERT INTO iteration_processes (
    iteration_id, process_info_id
) VALUES (
    $1, $2
)
RETURNING *;

-- name: GetProcessesByIterationID :many
SELECT pi.* FROM process_info pi
INNER JOIN iteration_processes ip ON pi.id = ip.process_info_id
WHERE ip.iteration_id = $1
ORDER BY pi.process_id;

-- name: GetIterationsByProcessInfoID :many
SELECT pih.* FROM process_iteration_history pih
INNER JOIN iteration_processes ip ON pih.id = ip.iteration_id
WHERE ip.process_info_id = $1
ORDER BY pih.created_at DESC;

-- ============================================================================
-- Process Query History Queries
-- ============================================================================

-- name: CreateProcessQueryHistory :one
INSERT INTO process_query_history (
    user_id, webhook_url, requested_pid, process_info_id, success, error_message
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetProcessQueryHistoryByUser :many
SELECT * FROM process_query_history
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetProcessQueryHistory :one
SELECT * FROM process_query_history
WHERE id = $1
LIMIT 1;

-- name: GetAllProcessQueryHistory :many
SELECT * FROM process_query_history
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetProcessQueryHistoryByUserPaginated :many
SELECT * FROM process_query_history
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetProcessQueryHistoryByPID :many
SELECT pqh.*, pi.* FROM process_query_history pqh
LEFT JOIN process_info pi ON pqh.process_info_id = pi.id
WHERE pqh.user_id = $1 AND pqh.requested_pid = $2
ORDER BY pqh.created_at DESC;

-- ============================================================================
-- Statistics and Analytics Queries
-- ============================================================================

-- name: GetUserProcessCount :one
SELECT COUNT(*) FROM process_info
WHERE user_id = $1;

-- name: GetUserIterationCount :one
SELECT COUNT(*) FROM process_iteration_history
WHERE user_id = $1;

-- name: GetUserQueryCount :one
SELECT COUNT(*) FROM process_query_history
WHERE user_id = $1;

-- name: GetMostQueriedProcesses :many
SELECT process_id, process_name, COUNT(*) as query_count
FROM process_info
WHERE user_id = $1
GROUP BY process_id, process_name
ORDER BY query_count DESC
LIMIT $2;
