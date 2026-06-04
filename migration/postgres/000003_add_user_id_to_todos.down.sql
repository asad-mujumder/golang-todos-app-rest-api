DROP INDEX IF EXISTS "idx_todos_user_id";

ALTER TABLE "todos"
  DROP COLUMN "user_id";