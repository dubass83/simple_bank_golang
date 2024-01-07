ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "owner_carrency_key";

ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";

DROP TABLE IF EXISTS "users";