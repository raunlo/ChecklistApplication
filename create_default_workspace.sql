-- Creates a default "Personal" workspace for every user who doesn't have one.
-- Safe to run multiple times — skips users who already have a default workspace.

DO $$
DECLARE
    v_user_id VARCHAR(255);
    v_workspace_id BIGINT;
BEGIN
    FOR v_user_id IN
        SELECT u.user_id FROM app_user u
        WHERE NOT EXISTS (
            SELECT 1 FROM workspace w
            WHERE w.owner_user_id = u.user_id AND w.is_default = TRUE
        )
    LOOP
        INSERT INTO workspace (owner_user_id, name, description, is_default, created_at, updated_at)
        VALUES (v_user_id, 'Personal', NULL, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id INTO v_workspace_id;

        INSERT INTO workspace_member (workspace_id, user_id, joined_at)
        VALUES (v_workspace_id, v_user_id, CURRENT_TIMESTAMP);

        RAISE NOTICE 'Created default workspace (id=%) for user %.', v_workspace_id, v_user_id;
    END LOOP;
END $$;
