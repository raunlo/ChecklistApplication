-- Migration script for existing production database
-- Run these statements in order. Each is idempotent (safe to run multiple times).
-- Last updated: 2025-05-02

-- ─────────────────────────────────────────────
-- 1. Rename is_personal → is_default
-- ─────────────────────────────────────────────
ALTER TABLE workspace RENAME COLUMN is_personal TO is_default;

-- ─────────────────────────────────────────────
-- 2. Unique constraint: one name per owner
-- ─────────────────────────────────────────────
ALTER TABLE workspace
    ADD CONSTRAINT uq_workspace_name_per_user UNIQUE (owner_user_id, name);

-- ─────────────────────────────────────────────
-- 3. Partial unique index: one default workspace per user
-- ─────────────────────────────────────────────
CREATE UNIQUE INDEX IF NOT EXISTS uq_workspace_default_per_user
    ON workspace(owner_user_id)
    WHERE is_default = TRUE;

-- ─────────────────────────────────────────────
-- 4. Rename existing default workspaces to "Personal"
--    (they were named after the user_id previously)
-- ─────────────────────────────────────────────
UPDATE workspace
SET name = 'Personal'
WHERE is_default = TRUE
  AND name != 'Personal';

-- ─────────────────────────────────────────────
-- 5. template_workspace join table
--    (replaces the scalar workspace_id column on TEMPLATE)
-- ─────────────────────────────────────────────
CREATE SEQUENCE IF NOT EXISTS template_share_id_sequence START 1 INCREMENT 1;

CREATE TABLE IF NOT EXISTS template_workspace (
    template_id  BIGINT NOT NULL REFERENCES TEMPLATE(ID) ON DELETE CASCADE,
    workspace_id BIGINT NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    assigned_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (template_id, workspace_id)
);

CREATE INDEX IF NOT EXISTS idx_tw_workspace ON template_workspace(workspace_id);

-- Migrate existing scalar assignments into the join table
INSERT INTO template_workspace (template_id, workspace_id)
SELECT id, workspace_id FROM TEMPLATE WHERE workspace_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- Drop old scalar column
DROP INDEX IF EXISTS idx_template_workspace;
ALTER TABLE TEMPLATE DROP COLUMN IF EXISTS workspace_id;

-- ─────────────────────────────────────────────
-- 6. job_lock table (background job coordination)
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS job_lock (
    job_name VARCHAR(100) PRIMARY KEY,
    last_run_at TIMESTAMP NOT NULL,
    locked_by VARCHAR(255) NULL,
    locked_at TIMESTAMP NULL
);

INSERT INTO job_lock (job_name, last_run_at)
VALUES ('soft_delete_cleanup', '1970-01-01 00:00:00')
ON CONFLICT (job_name) DO NOTHING;
