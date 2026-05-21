# migrate script — usage examples

All examples assume the scripts are in `scripts/` and you're running from the
project root. The `.env` and `migrations/` folder are also in the project root.

---

## Basic — default .env, default migrations dir

```bash
./scripts/migrate.sh up
./scripts/migrate.sh down
./scripts/migrate.sh down 3
./scripts/migrate.sh create --name add_users_table
./scripts/migrate.sh create                          # auto-name: migration_20260520_143022
./scripts/migrate.sh force 5
```

```powershell
.\scripts\migrate.ps1 up
.\scripts\migrate.ps1 down
.\scripts\migrate.ps1 down -Count 3
.\scripts\migrate.ps1 create -MigrationName add_users_table
.\scripts\migrate.ps1 create                         # auto-name: migration_20260520_143022
.\scripts\migrate.ps1 force -Version 5
```

---

## Custom .env file

```bash
./scripts/migrate.sh --env .env.production up
./scripts/migrate.sh --env .env.staging down 1
./scripts/migrate.sh --env /etc/myapp/.env create --name add_payments_table
```

```powershell
.\scripts\migrate.ps1 -Env .env.production up
.\scripts\migrate.ps1 -Env .env.staging down -Count 1
.\scripts\migrate.ps1 -Env C:\myapp\.env create -MigrationName add_payments_table
```

---

## Custom migrations directory

```bash
# Migrations live in db/migrations/ instead of migrations/
./scripts/migrate.sh --dir ./db/migrations up
./scripts/migrate.sh --dir ./db/migrations down 2
./scripts/migrate.sh --dir ./db/migrations create --name add_orders_table
./scripts/migrate.sh --dir ./db/migrations force 5
```

```powershell
.\scripts\migrate.ps1 -Dir ./db/migrations up
.\scripts\migrate.ps1 -Dir ./db/migrations down -Count 2
.\scripts\migrate.ps1 -Dir ./db/migrations create -MigrationName add_orders_table
.\scripts\migrate.ps1 -Dir ./db/migrations force -Version 5
```

---

## Combine custom .env and custom migrations dir

```bash
./scripts/migrate.sh --env .env.staging --dir ./db/migrations up
./scripts/migrate.sh --env .env.production --dir ./db/migrations down 1
```

```powershell
.\scripts\migrate.ps1 -Env .env.staging -Dir .\db\migrations up
.\scripts\migrate.ps1 -Env .env.production -Dir .\db\migrations down -Count 1
```

---

## Pass a full DSN directly (--url / -Url)

```bash
# Postgres
./scripts/migrate.sh --url "postgres://alice:s3cr3t@localhost:5432/shop?sslmode=disable" up

# MySQL
./scripts/migrate.sh --url "mysql://alice:s3cr3t@tcp(localhost:3306)/shop" up

# SQLite (no user/host needed — DB_NAME is the file path)
./scripts/migrate.sh --url "sqlite3://./data/app.db" up

# SQL Server
./scripts/migrate.sh --url "sqlserver://alice:s3cr3t@localhost:1433?database=shop" up

# Cassandra
./scripts/migrate.sh --url "cassandra://localhost:9042/shop" up

# ClickHouse
./scripts/migrate.sh --url "clickhouse://localhost:9000/shop" up

# With a custom migrations dir
./scripts/migrate.sh --url "postgres://alice:s3cr3t@localhost:5432/shop" \
  --dir ./db/migrations up
```

```powershell
# Postgres
.\scripts\migrate.ps1 -Url "postgres://alice:s3cr3t@localhost:5432/shop?sslmode=disable" up

# MySQL
.\scripts\migrate.ps1 -Url "mysql://alice:s3cr3t@tcp(localhost:3306)/shop" up

# SQLite
.\scripts\migrate.ps1 -Url "sqlite3://./data/app.db" up

# SQL Server
.\scripts\migrate.ps1 -Url "sqlserver://alice:s3cr3t@localhost:1433?database=shop" up

# With a custom migrations dir
.\scripts\migrate.ps1 -Url "postgres://alice:s3cr3t@localhost:5432/shop" `
  -Dir .\db\migrations up
```

---

## Read DSN from a named env variable (--url-var / -UrlVar)

Given a `.env` that contains any of these:

```dotenv
DATABASE_URL=postgres://alice:s3cr3t@localhost:5432/shop?sslmode=disable
PROD_DB_URL=postgres://alice:s3cr3t@prod.db.internal:5432/shop?sslmode=require
CI_DATABASE_URL=postgres://ci_user:ci_pass@ci-db:5432/shop_test
```

```bash
# Explicit var name — any variable works
./scripts/migrate.sh --url-var DATABASE_URL up
./scripts/migrate.sh --url-var PROD_DB_URL up
./scripts/migrate.sh --url-var CI_DATABASE_URL up

# DATABASE_URL is also auto-detected — no flag needed
./scripts/migrate.sh up
```

```powershell
.\scripts\migrate.ps1 -UrlVar DATABASE_URL up
.\scripts\migrate.ps1 -UrlVar PROD_DB_URL up
.\scripts\migrate.ps1 -UrlVar CI_DATABASE_URL up

# Auto-detected
.\scripts\migrate.ps1 up
```

---

## Override individual DB parts with direct values

Useful for targeting a different host/port without changing `.env`:

```bash
./scripts/migrate.sh --host prod.db.internal up
./scripts/migrate.sh --host staging.db.internal --port 5433 up
./scripts/migrate.sh --host localhost --port 5433 --db-name shop_test up
./scripts/migrate.sh --driver mysql --host localhost --user root --db-name shop up
```

```powershell
.\scripts\migrate.ps1 -DbHost prod.db.internal up
.\scripts\migrate.ps1 -DbHost staging.db.internal -Port 5433 up
.\scripts\migrate.ps1 -DbHost localhost -Port 5433 -DbName shop_test up
.\scripts\migrate.ps1 -Driver mysql -DbHost localhost -User root -DbName shop up
```

---

## Override individual parts with env var names (--*-var / -*Var)

Useful when your environment uses non-standard variable names
(e.g. Heroku uses `DATABASE_URL`, Render uses `RENDER_DB_URL`, psql uses `PGHOST`):

```bash
# psql standard env vars
./scripts/migrate.sh \
  --host-var     PGHOST     \
  --port-var     PGPORT     \
  --user-var     PGUSER     \
  --password-var PGPASSWORD \
  --db-name-var  PGDATABASE \
  up

# Railway-style vars
./scripts/migrate.sh \
  --host-var    RAILWAY_DB_HOST \
  --port-var    RAILWAY_DB_PORT \
  --db-name-var RAILWAY_DB_NAME \
  up
```

```powershell
# psql standard env vars
.\scripts\migrate.ps1 `
  -HostVar     PGHOST     `
  -PortVar     PGPORT     `
  -UserVar     PGUSER     `
  -PasswordVar PGPASSWORD `
  -DbNameVar   PGDATABASE `
  up

# Railway-style vars
.\scripts\migrate.ps1 `
  -HostVar    RAILWAY_DB_HOST `
  -PortVar    RAILWAY_DB_PORT `
  -DbNameVar  RAILWAY_DB_NAME `
  up
```

---

## Real-world scenarios

```bash
# Fresh repo: create first migration (no .env needed)
./scripts/migrate.sh create --name init_schema

# Local dev: standard .env in project root
./scripts/migrate.sh up

# Local dev: migrations in a non-default folder
./scripts/migrate.sh --dir ./db/migrations up

# CI pipeline: DSN injected as a secret env var
./scripts/migrate.sh --url-var CI_DATABASE_URL up

# CI with non-default migrations dir
./scripts/migrate.sh --url-var CI_DATABASE_URL --dir ./db/migrations up

# Staging deploy
./scripts/migrate.sh --env .env.staging --dir ./db/migrations up

# Oops: roll back the last migration in production
./scripts/migrate.sh --env .env.production --dir ./db/migrations down 1

# Schema got corrupted: force version back to known-good state
./scripts/migrate.sh --env .env.production --dir ./db/migrations force 12

# Test with a one-off DSN without changing .env
./scripts/migrate.sh --url "mysql://root@tcp(localhost:3306)/shop_test" \
  --dir ./db/migrations up
```

```powershell
# Fresh repo: create first migration
.\scripts\migrate.ps1 create -MigrationName init_schema

# Local dev
.\scripts\migrate.ps1 up

# Local dev with non-default migrations dir
.\scripts\migrate.ps1 -Dir .\db\migrations up

# CI pipeline
.\scripts\migrate.ps1 -UrlVar CI_DATABASE_URL up

# CI with non-default migrations dir
.\scripts\migrate.ps1 -UrlVar CI_DATABASE_URL -Dir ./db/migrations up

# Staging deploy
.\scripts\migrate.ps1 -Env .env.staging -Dir ./db/migrations up

# Roll back one migration in production
.\scripts\migrate.ps1 -Env .env.production -Dir ./db/migrations down -Count 1

# Force version
.\scripts\migrate.ps1 -Env .env.production -Dir ./db/migrations force -Version 12

# One-off DSN
.\scripts\migrate.ps1 -Url "mysql://root@tcp(localhost:3306)/shop_test" `
  -Dir .\db\migrations up
```

---

## DSN priority in action

```bash
# .env has DATABASE_URL and DB_HOST set — --url always wins
./scripts/migrate.sh --url "postgres://override@localhost/other" up

# .env has DATABASE_URL — auto-detected, no flag needed
./scripts/migrate.sh up

# .env has DATABASE_URL but you want a different var — --url-var wins over auto-detect
./scripts/migrate.sh --url-var PROD_DB_URL up

# .env has only DB_* parts — assembled automatically from DB_DRIVER + DB_HOST + …
./scripts/migrate.sh up

# Individual override beats the corresponding DB_* var but loses to --url / --url-var
./scripts/migrate.sh --host staging.internal up   # DB_HOST ignored, staging.internal used
```
