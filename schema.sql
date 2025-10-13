-- Schema para o exemplo
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Schema for process information based on webhook_handler.go ProcessInfo struct
CREATE TABLE process_info (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, -- NULL if no JWT token
    
    -- Basic process information
    process_id INTEGER NOT NULL,
    parent_process_id INTEGER NOT NULL,
    process_name VARCHAR(255) NOT NULL,
    thread_count INTEGER NOT NULL,
    handle_count INTEGER NOT NULL,
    base_priority INTEGER NOT NULL,
    
    -- Time information (stored as TEXT to match webhook response format)
    create_time TEXT NOT NULL,
    user_time TEXT NOT NULL,
    kernel_time TEXT NOT NULL,
    
    -- Memory information (stored as TEXT to match webhook response format)
    working_set_size TEXT NOT NULL,
    peak_working_set_size TEXT NOT NULL,
    virtual_size TEXT NOT NULL,
    peak_virtual_size TEXT NOT NULL,
    pagefile_usage TEXT NOT NULL,
    peak_pagefile_usage TEXT NOT NULL,
    page_fault_count INTEGER NOT NULL,
    
    -- I/O information
    read_operation_count BIGINT NOT NULL,
    write_operation_count BIGINT NOT NULL,
    other_operation_count BIGINT NOT NULL,
    read_transfer_count BIGINT NOT NULL,
    write_transfer_count BIGINT NOT NULL,
    other_transfer_count BIGINT NOT NULL,
    
    -- Process address
    current_process_address TEXT NOT NULL,
    
    -- Next process information
    next_process_eprocess_address TEXT,
    next_process_name VARCHAR(255),
    next_process_id INTEGER,
    
    -- Previous process information
    previous_process_eprocess_address TEXT,
    previous_process_name VARCHAR(255),
    previous_process_id INTEGER,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Index for faster queries
    CONSTRAINT unique_process_snapshot UNIQUE (process_id, current_process_address, created_at)
);

-- Table to track iterate-processes history
CREATE TABLE process_iteration_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, -- NULL if no JWT token
    webhook_url TEXT NOT NULL,
    process_count INTEGER NOT NULL,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table to link process_info to iteration history (many-to-many)
CREATE TABLE iteration_processes (
    id BIGSERIAL PRIMARY KEY,
    iteration_id BIGINT NOT NULL REFERENCES process_iteration_history(id) ON DELETE CASCADE,
    process_info_id BIGINT NOT NULL REFERENCES process_info(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT unique_iteration_process UNIQUE (iteration_id, process_info_id)
);

-- Table to track individual process queries (processByPid)
CREATE TABLE process_query_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, -- NULL if no JWT token
    webhook_url TEXT NOT NULL,
    requested_pid INTEGER NOT NULL,
    process_info_id BIGINT REFERENCES process_info(id) ON DELETE SET NULL,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX idx_process_info_user_id ON process_info(user_id);
CREATE INDEX idx_process_info_process_id ON process_info(process_id);
CREATE INDEX idx_process_info_created_at ON process_info(created_at DESC);
CREATE INDEX idx_process_iteration_history_user_id ON process_iteration_history(user_id);
CREATE INDEX idx_process_iteration_history_created_at ON process_iteration_history(created_at DESC);
CREATE INDEX idx_process_query_history_user_id ON process_query_history(user_id);
CREATE INDEX idx_process_query_history_created_at ON process_query_history(created_at DESC);
CREATE INDEX idx_iteration_processes_iteration_id ON iteration_processes(iteration_id);
