-- Migration to update from old structure to new snapshot-based structure
-- Run this migration if you have existing data

BEGIN;

-- Step 1: Create new tables
CREATE TABLE IF NOT EXISTS process_snapshots (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    webhook_url TEXT NOT NULL,
    snapshot_type VARCHAR(50) NOT NULL,
    process_count INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Step 2: Migrate data from process_iteration_history to process_snapshots
INSERT INTO process_snapshots (id, user_id, webhook_url, snapshot_type, process_count, success, error_message, created_at, updated_at)
SELECT 
    id,
    user_id,
    webhook_url,
    'iteration' as snapshot_type,
    process_count,
    success,
    error_message,
    created_at,
    created_at as updated_at
FROM process_iteration_history
WHERE EXISTS (SELECT 1 FROM process_iteration_history);

-- Step 3: Add snapshot_id column to process_info if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='process_info' AND column_name='snapshot_id') THEN
        ALTER TABLE process_info ADD COLUMN snapshot_id BIGINT;
    END IF;
END $$;

-- Step 4: Migrate process_info data to link with snapshots
-- Link processes to their snapshots via iteration_processes table
UPDATE process_info pi
SET snapshot_id = ip.iteration_id
FROM iteration_processes ip
WHERE pi.id = ip.process_info_id
AND pi.snapshot_id IS NULL;

-- Step 5: For processes without a snapshot, create individual snapshots
DO $$
DECLARE
    process_record RECORD;
    new_snapshot_id BIGINT;
BEGIN
    FOR process_record IN 
        SELECT * FROM process_info WHERE snapshot_id IS NULL
    LOOP
        INSERT INTO process_snapshots (user_id, webhook_url, snapshot_type, process_count, success, created_at, updated_at)
        VALUES (
            process_record.user_id,
            'migrated',
            'query',
            1,
            true,
            process_record.created_at,
            process_record.created_at
        )
        RETURNING id INTO new_snapshot_id;
        
        UPDATE process_info 
        SET snapshot_id = new_snapshot_id 
        WHERE id = process_record.id;
    END LOOP;
END $$;

-- Step 6: Make snapshot_id NOT NULL after migration
ALTER TABLE process_info ALTER COLUMN snapshot_id SET NOT NULL;

-- Step 7: Add foreign key constraint
ALTER TABLE process_info 
ADD CONSTRAINT fk_process_info_snapshot 
FOREIGN KEY (snapshot_id) REFERENCES process_snapshots(id) ON DELETE CASCADE;

-- Step 8: Drop old unique constraint and add new one
ALTER TABLE process_info DROP CONSTRAINT IF EXISTS unique_process_snapshot;
ALTER TABLE process_info 
ADD CONSTRAINT unique_process_in_snapshot 
UNIQUE (snapshot_id, process_id, current_process_address);

-- Step 9: Create process_queries table
CREATE TABLE IF NOT EXISTS process_queries (
    id BIGSERIAL PRIMARY KEY,
    snapshot_id BIGINT NOT NULL REFERENCES process_snapshots(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    webhook_url TEXT NOT NULL,
    requested_pid INTEGER NOT NULL,
    process_info_id BIGINT REFERENCES process_info(id) ON DELETE SET NULL,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Step 10: Migrate data from process_query_history to process_queries
-- First, create snapshots for each query
INSERT INTO process_snapshots (user_id, webhook_url, snapshot_type, process_count, success, error_message, created_at, updated_at)
SELECT 
    user_id,
    webhook_url,
    'query' as snapshot_type,
    1 as process_count,
    success,
    error_message,
    created_at,
    created_at as updated_at
FROM process_query_history
WHERE EXISTS (SELECT 1 FROM process_query_history);

-- Then migrate the queries
INSERT INTO process_queries (snapshot_id, user_id, webhook_url, requested_pid, process_info_id, success, error_message, created_at)
SELECT 
    ps.id as snapshot_id,
    pqh.user_id,
    pqh.webhook_url,
    pqh.requested_pid,
    pqh.process_info_id,
    pqh.success,
    pqh.error_message,
    pqh.created_at
FROM process_query_history pqh
JOIN process_snapshots ps ON 
    ps.user_id = pqh.user_id AND 
    ps.webhook_url = pqh.webhook_url AND 
    ps.snapshot_type = 'query' AND
    ps.created_at = pqh.created_at;

-- Step 11: Create new indexes
CREATE INDEX IF NOT EXISTS idx_process_snapshots_user_id ON process_snapshots(user_id);
CREATE INDEX IF NOT EXISTS idx_process_snapshots_created_at ON process_snapshots(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_process_snapshots_type ON process_snapshots(snapshot_type);

CREATE INDEX IF NOT EXISTS idx_process_info_snapshot_id ON process_info(snapshot_id);

CREATE INDEX IF NOT EXISTS idx_process_queries_snapshot_id ON process_queries(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_process_queries_user_id ON process_queries(user_id);
CREATE INDEX IF NOT EXISTS idx_process_queries_created_at ON process_queries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_process_queries_requested_pid ON process_queries(requested_pid);

-- Step 12: Drop old tables (uncomment when ready)
-- DROP TABLE IF EXISTS iteration_processes CASCADE;
-- DROP TABLE IF EXISTS process_query_history CASCADE;
-- DROP TABLE IF EXISTS process_iteration_history CASCADE;

COMMIT;

-- To rollback if needed:
-- ROLLBACK;
