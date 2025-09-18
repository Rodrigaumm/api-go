-- name: GetUsers :many
SELECT id, name, email, password, created_at, updated_at FROM users
ORDER BY id;

-- name: GetUser :one
SELECT id, name, email, password, created_at, updated_at FROM users
WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
  name, email, password
) VALUES (
  $1, $2, $3
)
RETURNING id, name, email, password, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
  set name = $2,
  email = $3,
  password = $4,
  updated_at = NOW()
WHERE id = $1
RETURNING id, name, email, password, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, name, email, password, created_at, updated_at FROM users
WHERE email = $1 LIMIT 1;

-- Process Info Queries

-- name: CreateProcessInfo :one
INSERT INTO process_info (
    user_id, process_id, parent_process_id, process_name, thread_count, handle_count,
    base_priority, create_time, user_time, kernel_time, working_set_size, peak_working_set_size,
    virtual_size, peak_virtual_size, pagefile_usage, peak_pagefile_usage, page_fault_count,
    read_operation_count, write_operation_count, other_operation_count, read_transfer_count,
    write_transfer_count, other_transfer_count, current_process_address, next_process_address,
    previous_process_address
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
)
RETURNING id, user_id, process_id, parent_process_id, process_name, thread_count, handle_count,
    base_priority, create_time, user_time, kernel_time, working_set_size, peak_working_set_size,
    virtual_size, peak_virtual_size, pagefile_usage, peak_pagefile_usage, page_fault_count,
    read_operation_count, write_operation_count, other_operation_count, read_transfer_count,
    write_transfer_count, other_transfer_count, current_process_address, next_process_address,
    previous_process_address, created_at, updated_at;

-- name: GetProcessInfosByUser :many
SELECT id, user_id, process_id, parent_process_id, process_name, thread_count, handle_count,
    base_priority, create_time, user_time, kernel_time, working_set_size, peak_working_set_size,
    virtual_size, peak_virtual_size, pagefile_usage, peak_pagefile_usage, page_fault_count,
    read_operation_count, write_operation_count, other_operation_count, read_transfer_count,
    write_transfer_count, other_transfer_count, current_process_address, next_process_address,
    previous_process_address, created_at, updated_at
FROM process_info
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetProcessInfo :one
SELECT id, user_id, process_id, parent_process_id, process_name, thread_count, handle_count,
    base_priority, create_time, user_time, kernel_time, working_set_size, peak_working_set_size,
    virtual_size, peak_virtual_size, pagefile_usage, peak_pagefile_usage, page_fault_count,
    read_operation_count, write_operation_count, other_operation_count, read_transfer_count,
    write_transfer_count, other_transfer_count, current_process_address, next_process_address,
    previous_process_address, created_at, updated_at
FROM process_info
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: DeleteProcessInfo :exec
DELETE FROM process_info
WHERE id = $1 AND user_id = $2;

-- name: GetProcessInfosByProcessId :many
SELECT id, user_id, process_id, parent_process_id, process_name, thread_count, handle_count,
    base_priority, create_time, user_time, kernel_time, working_set_size, peak_working_set_size,
    virtual_size, peak_virtual_size, pagefile_usage, peak_pagefile_usage, page_fault_count,
    read_operation_count, write_operation_count, other_operation_count, read_transfer_count,
    write_transfer_count, other_transfer_count, current_process_address, next_process_address,
    previous_process_address, created_at, updated_at
FROM process_info
WHERE user_id = $1 AND process_id = $2
ORDER BY created_at DESC;