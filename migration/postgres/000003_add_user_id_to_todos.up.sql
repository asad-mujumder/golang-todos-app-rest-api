ALTER TABLE "todos"
  ADD COLUMN "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE;

CREATE INDEX "idx_todos_user_id" ON "todos"("user_id");