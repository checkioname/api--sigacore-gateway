-- Remove índice
DROP INDEX IF EXISTS "idx_users_whitelisted";

-- Remove coluna de whitelist
ALTER TABLE "users" DROP COLUMN IF EXISTS "is_whitelisted"; 