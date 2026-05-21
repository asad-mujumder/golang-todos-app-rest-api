#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────────────
# migrate.sh — Database migration helper (Linux/macOS)
#
# USAGE
#   ./scripts/migrate.sh [OPTIONS] <command>
#
# COMMANDS
#   up                       Run all pending migrations
#   down [count]             Roll back N migrations (default: 1)
#   create [--name <name>]   Create a new migration file
#   force <version>          Force the migration version number
#
# DSN RESOLUTION ORDER (first match wins)
#   1. --url <dsn>           Literal DSN string
#   2. --url-var <VAR>       Named env var holding the DSN
#   3. DATABASE_URL          Auto-detected from .env
#   4. Built from parts      DB_DRIVER + DB_HOST + DB_PORT + …
#
# SUPPORTED DRIVERS (DB_DRIVER)
#   postgres | postgresql | cockroachdb | redshift
#   mysql | mariadb
#   sqlite3 | sqlite
#   sqlserver | mssql
#   cassandra | clickhouse
# ──────────────────────────────────────────────────────────────────────────────

set -euo pipefail

# ── LOCATE PROJECT ROOT ───────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# ── FLAG DEFAULTS ─────────────────────────────────────────────────────────────
OPT_ENV_FILE=""
OPT_DIR=""               # --dir : migrations folder; default $PROJECT_ROOT/migrations
OPT_NAME=""              # --name: migration name (create command only)

OPT_URL=""
OPT_URL_VAR=""

OPT_DRIVER=""     OPT_DRIVER_VAR=""
OPT_HOST=""       OPT_HOST_VAR=""
OPT_PORT=""       OPT_PORT_VAR=""
OPT_USER=""       OPT_USER_VAR=""
OPT_PASSWORD=""   OPT_PASSWORD_VAR=""
OPT_DBNAME=""     OPT_DBNAME_VAR=""
OPT_SSLMODE=""    OPT_SSLMODE_VAR=""

# ── HELPERS ───────────────────────────────────────────────────────────────────

info()    { printf '\033[1;32m[INFO]\033[0m  %s\n' "$*"; }
warning() { printf '\033[1;33m[WARN]\033[0m  %s\n' "$*"; }
error()   { printf '\033[1;31m[ERROR]\033[0m %s\n' "$*" >&2; }

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS] <command>

COMMANDS
  up                        Run all pending migrations
  down [count]              Roll back N migrations (default: 1)
  create [--name <name>]    Create a new migration file
                            If --name is omitted, a timestamped name is used:
                            migration_YYYYMMDD_HHMMSS
  force <version>           Force set the migration version

GENERAL OPTIONS
  --env  <path>   Path to .env file (default: $PROJECT_ROOT/.env)
  --dir  <path>   Migrations folder   (default: $PROJECT_ROOT/migrations)

DSN OPTIONS  (first match wins)
  --url     <dsn>   Use this connection string directly
  --url-var <VAR>   Read DSN from this env variable (e.g. DATABASE_URL)
  (auto)            DATABASE_URL detected in .env

PART OVERRIDES  (direct value takes priority over var name)
  --driver  <val>   --driver-var  <VAR>   default env var: DB_DRIVER
  --host    <val>   --host-var    <VAR>   default env var: DB_HOST
  --port    <val>   --port-var    <VAR>   default env var: DB_PORT
  --user    <val>   --user-var    <VAR>   default env var: DB_USER
  --password <val>  --password-var <VAR>  default env var: DB_PASSWORD
  --name    <val>   --name-var    <VAR>   default env var: DB_NAME
  --sslmode <val>   --sslmode-var <VAR>   default env var: DB_SSLMODE

EXAMPLES
  ./scripts/migrate.sh up
  ./scripts/migrate.sh down 3
  ./scripts/migrate.sh create --name add_users_table
  ./scripts/migrate.sh create                            # auto-named
  ./scripts/migrate.sh force 5
  ./scripts/migrate.sh --dir ./db/migrations up
  ./scripts/migrate.sh --env .env.staging --dir ./db/migrations up
  ./scripts/migrate.sh --url "postgres://user:pass@localhost:5432/db" up
  ./scripts/migrate.sh --url-var DATABASE_URL up
EOF
    exit 1
}

require_command() {
    if ! command -v "$1" &>/dev/null; then
        error "Command '$1' not found. Please install it first."
        exit 1
    fi
}

confirm() {
    read -r -p "$1 [y/N] " response
    [[ "${response,,}" == "y" ]]
}

# ── ARGUMENT PARSER ───────────────────────────────────────────────────────────
# Walks $@ and separates flags from positional args.
# Each --flag <value> pair shifts twice; positional args go into POSITIONAL_ARGS.

parse_args() {
    POSITIONAL_ARGS=()

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --env)          OPT_ENV_FILE="$2";      shift 2 ;;
            --env=*)        OPT_ENV_FILE="${1#--env=}"; shift ;;

            --dir)          OPT_DIR="$2";            shift 2 ;;
            --dir=*)        OPT_DIR="${1#--dir=}";   shift ;;

            # --name is shared: for 'create' it sets the migration name;
            # the DSN part overrides also use --name (for DB_NAME) via --name-var.
            # We resolve ambiguity in cmd_create: OPT_NAME is checked first.
            --name)         OPT_NAME="$2";           shift 2 ;;
            --name=*)       OPT_NAME="${1#--name=}"; shift ;;

            --url)          OPT_URL="$2";            shift 2 ;;
            --url=*)        OPT_URL="${1#--url=}";   shift ;;
            --url-var)      OPT_URL_VAR="$2";        shift 2 ;;
            --url-var=*)    OPT_URL_VAR="${1#--url-var=}"; shift ;;

            --driver)       OPT_DRIVER="$2";         shift 2 ;;
            --driver-var)   OPT_DRIVER_VAR="$2";     shift 2 ;;
            --host)         OPT_HOST="$2";           shift 2 ;;
            --host-var)     OPT_HOST_VAR="$2";       shift 2 ;;
            --port)         OPT_PORT="$2";           shift 2 ;;
            --port-var)     OPT_PORT_VAR="$2";       shift 2 ;;
            --user)         OPT_USER="$2";           shift 2 ;;
            --user-var)     OPT_USER_VAR="$2";       shift 2 ;;
            --password)     OPT_PASSWORD="$2";       shift 2 ;;
            --password-var) OPT_PASSWORD_VAR="$2";   shift 2 ;;
            --db-name)      OPT_DBNAME="$2";         shift 2 ;;
            --db-name-var)  OPT_DBNAME_VAR="$2";     shift 2 ;;
            --sslmode)      OPT_SSLMODE="$2";        shift 2 ;;
            --sslmode-var)  OPT_SSLMODE_VAR="$2";    shift 2 ;;

            -h|--help) usage ;;
            *) POSITIONAL_ARGS+=("$1"); shift ;;
        esac
    done
}

# ── LOAD .env ─────────────────────────────────────────────────────────────────

load_env() {
    local env_file="$1"

    if [[ ! -f "$env_file" ]]; then
        error ".env file not found: $env_file"
        exit 1
    fi

    info "Loading env from: $env_file"

    while IFS= read -r line || [[ -n "$line" ]]; do
        [[ -z "$line" || "$line" == \#* ]] && continue

        if [[ "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
            local key="${BASH_REMATCH[1]}"
            local value="${BASH_REMATCH[2]}"

            if [[ "$value" =~ ^\"([^\"]*)\" ]]; then
                value="${BASH_REMATCH[1]}"
            elif [[ "$value" =~ ^\'([^\']*)\' ]]; then
                value="${BASH_REMATCH[1]}"
            else
                value="${value%%[[:space:]]#*}"
                value="${value%"${value##*[! ]}"}"
            fi

            export "$key=$value"
        fi
    done < "$env_file"
}

# ── FIELD RESOLVER ────────────────────────────────────────────────────────────
# Priority: direct value → named env var → default DB_* env var

resolve_field() {
    local direct="$1" var_name="$2" default_var="$3"

    if [[ -n "$direct" ]];    then printf '%s' "$direct";          return; fi
    if [[ -n "$var_name" ]];  then printf '%s' "${!var_name:-}";   return; fi
    printf '%s' "${!default_var:-}"
}

# ── URL-ENCODE ────────────────────────────────────────────────────────────────

url_encode() {
    local s="$1"
    s="${s//%/%25}"; s="${s//@/%40}"; s="${s//:/%3A}"; s="${s//\//%2F}"
    s="${s//?/%3F}"; s="${s//=/%3D}"; s="${s//&/%26}"
    s="${s//+/%2B}"; s="${s//#/%23}"
    printf '%s' "$s"
}

# ── DSN BUILDER ───────────────────────────────────────────────────────────────

build_dsn() {
    local driver; driver="$(resolve_field "$OPT_DRIVER"   "$OPT_DRIVER_VAR"   "DB_DRIVER")"
    local host;   host="$(  resolve_field "$OPT_HOST"     "$OPT_HOST_VAR"     "DB_HOST")"
    local port;   port="$(  resolve_field "$OPT_PORT"     "$OPT_PORT_VAR"     "DB_PORT")"
    local user;   user="$(  resolve_field "$OPT_USER"     "$OPT_USER_VAR"     "DB_USER")"
    local pass_raw; pass_raw="$(resolve_field "$OPT_PASSWORD" "$OPT_PASSWORD_VAR" "DB_PASSWORD")"
    local name;   name="$(  resolve_field "$OPT_DBNAME"   "$OPT_DBNAME_VAR"   "DB_NAME")"
    local ssl;    ssl="$(   resolve_field "$OPT_SSLMODE"  "$OPT_SSLMODE_VAR"  "DB_SSLMODE")"
    ssl="${ssl:-disable}"

    if [[ -z "$driver" ]]; then
        error "DB driver not set. Use --driver, --driver-var, or DB_DRIVER in .env."
        exit 1
    fi

    local pass; pass="$(url_encode "$pass_raw")"
    local creds; creds="$([ -n "$pass_raw" ] && printf '%s:%s' "$user" "$pass" || printf '%s' "$user")"

    case "$driver" in
        sqlite3|sqlite) ;;
        *)
            [[ -z "$user" ]] && { error "DB user not set. Use --user / --user-var / DB_USER."; exit 1; }
            [[ -z "$host" ]] && { error "DB host not set. Use --host / --host-var / DB_HOST."; exit 1; }
            [[ -z "$name" ]] && { error "DB name not set. Use --db-name / --db-name-var / DB_NAME."; exit 1; }
            ;;
    esac

    case "$driver" in
        postgres|postgresql|cockroachdb|redshift)
            printf 'postgres://%s@%s:%s/%s?sslmode=%s' "$creds" "$host" "${port:-5432}" "$name" "$ssl" ;;
        mysql|mariadb)
            printf 'mysql://%s@tcp(%s:%s)/%s' "$creds" "$host" "${port:-3306}" "$name" ;;
        sqlite3|sqlite)
            [[ -z "$name" ]] && { error "DB_NAME must be the SQLite file path."; exit 1; }
            printf 'sqlite3://%s' "$name" ;;
        sqlserver|mssql)
            printf 'sqlserver://%s@%s:%s?database=%s' "$creds" "$host" "${port:-1433}" "$name" ;;
        cassandra)
            printf 'cassandra://%s:%s/%s' "$host" "${port:-9042}" "$name" ;;
        clickhouse)
            printf 'clickhouse://%s:%s/%s' "$host" "${port:-9000}" "$name" ;;
        *)
            error "Unknown driver '$driver'. Supported: postgres, cockroachdb, redshift,"
            error "  mysql, mariadb, sqlite3, sqlserver, mssql, cassandra, clickhouse"
            exit 1 ;;
    esac
}

# ── RESOLVE DSN ───────────────────────────────────────────────────────────────

resolve_dsn() {
    if [[ -n "$OPT_URL" ]]; then
        info "Using --url DSN."
        printf '%s' "$OPT_URL"; return
    fi

    if [[ -n "$OPT_URL_VAR" ]]; then
        local val="${!OPT_URL_VAR:-}"
        [[ -z "$val" ]] && { error "--url-var '$OPT_URL_VAR' is empty or not set."; exit 1; }
        info "Using DSN from \$$OPT_URL_VAR."
        printf '%s' "$val"; return
    fi

    if [[ -n "${DATABASE_URL:-}" ]]; then
        info "Auto-detected DATABASE_URL."
        printf '%s' "$DATABASE_URL"; return
    fi

    build_dsn
}

# ── COMMANDS ──────────────────────────────────────────────────────────────────

cmd_up() {
    local dsn="$1" dir="$2"
    info "Running all pending migrations..."
    migrate -path "$dir" -database "$dsn" up
    info "Done."
}

cmd_down() {
    local dsn="$1" dir="$2" count="${3:-1}"
    if ! [[ "$count" =~ ^[0-9]+$ ]]; then
        error "Count must be a positive integer, got: '$count'"; exit 1
    fi
    warning "About to roll back $count migration(s)."
    if confirm "Continue?"; then
        migrate -path "$dir" -database "$dsn" down "$count"
        info "Rolled back $count migration(s)."
    else
        info "Aborted."
    fi
}

cmd_create() {
    local dir="$1"

    # Use --name if provided, otherwise generate a timestamped default name.
    # 'date +...' formats the current date/time; the result looks like:
    #   migration_20260520_143022
    local name="${OPT_NAME:-migration_$(date +'%Y%m%d_%H%M%S')}"

    if [[ -z "$OPT_NAME" ]]; then
        info "No --name given. Using generated name: $name"
    fi

    migrate create -ext sql -dir "$dir" -seq "$name"
    info "Created migration: $name"
}

cmd_force() {
    local dsn="$1" dir="$2" version="$3"
    if [[ -z "$version" ]]; then
        error "Version required. Example: $(basename "$0") force 5"; exit 1
    fi
    warning "Forcing migration version to $version."
    migrate -path "$dir" -database "$dsn" force "$version"
    info "Version forced to $version."
}

# ── MAIN ──────────────────────────────────────────────────────────────────────

main() {
    require_command migrate
    parse_args "$@"
    set -- "${POSITIONAL_ARGS[@]+"${POSITIONAL_ARGS[@]}"}"
    [[ $# -lt 1 ]] && usage

    local env_file="${OPT_ENV_FILE:-$PROJECT_ROOT/.env}"
    load_env "$env_file"

    # Resolve the migrations directory.
    # --dir wins over the default; we expand ~ and resolve relative paths.
    local migrations_dir
    migrations_dir="$(realpath -m "${OPT_DIR:-$PROJECT_ROOT/migrations}")"
    info "Migrations dir: $migrations_dir"

    local command="$1"; shift

    # 'create' doesn't need a DSN — skip resolve_dsn to avoid validation errors
    # when DB vars aren't set (e.g. in a fresh repo with no .env yet).
    if [[ "$command" == "create" ]]; then
        cmd_create "$migrations_dir"
        return
    fi

    local dsn; dsn="$(resolve_dsn)"

    case "$command" in
        up)    cmd_up    "$dsn" "$migrations_dir" ;;
        down)  cmd_down  "$dsn" "$migrations_dir" "${1:-}" ;;
        force) cmd_force "$dsn" "$migrations_dir" "${1:-}" ;;
        *)     error "Unknown command: '$command'"; usage ;;
    esac
}

main "$@"