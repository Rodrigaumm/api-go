-- Schema para o exemplo
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Table to represent a "snapshot" or "session" of process capture
-- Each call to iterate-processes creates a new snapshot
CREATE TABLE process_snapshots (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, -- NULL if no JWT token
    webhook_url TEXT NOT NULL,
    snapshot_type VARCHAR(50) NOT NULL, -- 'iteration' or 'query'
    process_count INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Schema for process information based on webhook_handler.go ProcessInfo struct
CREATE TABLE process_info (
    id BIGSERIAL PRIMARY KEY,
    snapshot_id BIGINT NOT NULL REFERENCES process_snapshots(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, -- NULL if no JWT token
    
    -- Basic process information
    process_id BIGINT NOT NULL,
    parent_process_id BIGINT NOT NULL,
    process_name VARCHAR(255) NOT NULL,
    thread_count INTEGER NOT NULL,
    handle_count INTEGER NOT NULL,
    base_priority INTEGER NOT NULL,
    
    -- Time information (stored as TEXT to match webhook response format)
    create_time TEXT NOT NULL,
    user_time INTEGER NOT NULL,
    kernel_time INTEGER NOT NULL,
    
    -- Memory information (stored as TEXT to match webhook response format)
    working_set_size BIGINT NOT NULL,
    peak_working_set_size BIGINT NOT NULL,
    virtual_size BIGINT NOT NULL,
    peak_virtual_size BIGINT NOT NULL,
    
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
    next_process_id BIGINT,
    
    -- Previous process information
    previous_process_eprocess_address TEXT,
    previous_process_name VARCHAR(255),
    previous_process_id BIGINT,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Index for faster queries
    CONSTRAINT unique_process_in_snapshot UNIQUE (snapshot_id, process_id, current_process_address)
);

-- Table to track individual process queries by PID
-- Each query can either create a new snapshot or add to an existing one
CREATE TABLE process_queries (
    id BIGSERIAL PRIMARY KEY,
    snapshot_id BIGINT NOT NULL REFERENCES process_snapshots(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, -- NULL if no JWT token
    webhook_url TEXT NOT NULL,
    requested_pid INTEGER NOT NULL,
    process_info_id BIGINT REFERENCES process_info(id) ON DELETE SET NULL,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX idx_process_snapshots_user_id ON process_snapshots(user_id);
CREATE INDEX idx_process_snapshots_created_at ON process_snapshots(created_at DESC);
CREATE INDEX idx_process_snapshots_type ON process_snapshots(snapshot_type);

CREATE INDEX idx_process_info_snapshot_id ON process_info(snapshot_id);
CREATE INDEX idx_process_info_user_id ON process_info(user_id);
CREATE INDEX idx_process_info_process_id ON process_info(process_id);
CREATE INDEX idx_process_info_created_at ON process_info(created_at DESC);

CREATE INDEX idx_process_queries_snapshot_id ON process_queries(snapshot_id);
CREATE INDEX idx_process_queries_user_id ON process_queries(user_id);
CREATE INDEX idx_process_queries_created_at ON process_queries(created_at DESC);
CREATE INDEX idx_process_queries_requested_pid ON process_queries(requested_pid);
