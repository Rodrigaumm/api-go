-- Schema para o exemplo
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Schema for process information
CREATE TABLE process_info (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Basic process information
    process_id BIGINT NOT NULL,
    parent_process_id BIGINT NOT NULL,
    process_name VARCHAR(64) NOT NULL,
    thread_count BIGINT NOT NULL,
    handle_count BIGINT NOT NULL,
    base_priority INTEGER NOT NULL,
    create_time TIMESTAMP NOT NULL,
    user_time BIGINT NOT NULL,
    kernel_time BIGINT NOT NULL,

    -- Memory information
    working_set_size BIGINT NOT NULL,
    peak_working_set_size BIGINT NOT NULL,
    virtual_size BIGINT NOT NULL,
    peak_virtual_size BIGINT NOT NULL,
    pagefile_usage BIGINT NOT NULL,
    peak_pagefile_usage BIGINT NOT NULL,
    page_fault_count BIGINT NOT NULL,

    -- I/O information
    read_operation_count BIGINT NOT NULL,
    write_operation_count BIGINT NOT NULL,
    other_operation_count BIGINT NOT NULL,
    read_transfer_count BIGINT NOT NULL,
    write_transfer_count BIGINT NOT NULL,
    other_transfer_count BIGINT NOT NULL,

    -- Process addresses (stored as text since they're pointers)
    current_process_address TEXT,
    next_process_address TEXT,
    previous_process_address TEXT,

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);