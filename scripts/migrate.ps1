# ──────────────────────────────────────────────────────────────────────────────
# migrate.ps1 — Database migration helper (Windows PowerShell)
#
# USAGE
#   .\scripts\migrate.ps1 [OPTIONS] <command>
#
# COMMANDS
#   up                          Run all pending migrations
#   down [-Count N]             Roll back N migrations (default: 1)
#   create [-MigrationName <n>] Create a new migration file
#                               If -MigrationName is omitted, a timestamped
#                               name is used: migration_YYYYMMDD_HHMMSS
#   force -Version <v>          Force the migration version number
#
# DSN RESOLUTION ORDER (first match wins)
#   1. -Url <dsn>               Literal DSN string
#   2. -UrlVar <VAR>            Named env var holding the DSN
#   3. DATABASE_URL             Auto-detected from .env
#   4. Built from parts         DB_DRIVER + DB_HOST + DB_PORT + …
#
# SUPPORTED DRIVERS (DB_DRIVER)
#   postgres | postgresql | cockroachdb | redshift
#   mysql | mariadb
#   sqlite3 | sqlite
#   sqlserver | mssql
#   cassandra | clickhouse
# ──────────────────────────────────────────────────────────────────────────────

[CmdletBinding()]
param(
    # ── Command ───────────────────────────────────────────────────────────────
    [Parameter(Position = 0)]
    [ValidateSet('up', 'down', 'create', 'force')]
    [string]$Command,

    # ── Command-specific args ─────────────────────────────────────────────────
    [string]$MigrationName = '',   # create -MigrationName add_users_table
    [string]$Count         = '',   # down   -Count 3
    [string]$Version       = '',   # force  -Version 5

    # ── General options ───────────────────────────────────────────────────────
    [string]$Env = '',    # -Env  : path to .env file
    [string]$Dir = '',    # -Dir  : migrations folder path

    # ── DSN shortcuts ─────────────────────────────────────────────────────────
    [string]$Url    = '',
    [string]$UrlVar = '',

    # ── Individual part: direct value ─────────────────────────────────────────
    [string]$Driver   = '',
    [string]$DbHost   = '',    # DbHost to avoid clash with $Host (PS built-in)
    [string]$Port     = '',
    [string]$User     = '',
    [string]$Password = '',
    [string]$DbName   = '',
    [string]$SslMode  = '',

    # ── Individual part: env var name to read ─────────────────────────────────
    [string]$DriverVar   = '',
    [string]$HostVar     = '',
    [string]$PortVar     = '',
    [string]$UserVar     = '',
    [string]$PasswordVar = '',
    [string]$DbNameVar   = '',
    [string]$SslModeVar  = ''
)

# --- STRICT MODE --------------------------------------------------------------
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# --- LOCATE PROJECT ROOT ------------------------------------------------------
$ScriptDir   = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir

$DefaultEnvPath        = Join-Path $ProjectRoot '.env'
$DefaultMigrationsPath = Join-Path $ProjectRoot 'migrations'

# --- HELPERS ------------------------------------------------------------------

function Write-Info { param([string]$Msg) Write-Host "[INFO]  $Msg" -ForegroundColor Green }
function Write-Warn { param([string]$Msg) Write-Host "[WARN]  $Msg" -ForegroundColor Yellow }
function Write-Err  { param([string]$Msg) Write-Host "[ERROR] $Msg" -ForegroundColor Red }

function Show-Usage {
    Write-Host @"
Usage: .\migrate.ps1 [OPTIONS] <command>

COMMANDS
  up                            Run all pending migrations
  down [-Count N]               Roll back N migrations (default: 1)
  create [-MigrationName <n>]   Create a new migration file
                                Omit -MigrationName to use a generated name:
                                migration_YYYYMMDD_HHMMSS
  force -Version <v>            Force set the migration version

GENERAL OPTIONS
  -Env <path>     Path to .env file      (default: $DefaultEnvPath)
  -Dir <path>     Migrations folder path (default: $DefaultMigrationsPath)

DSN OPTIONS  (first match wins)
  -Url    <dsn>   Use this connection string directly
  -UrlVar <VAR>   Read DSN from this env variable (e.g. DATABASE_URL)
  (auto)          DATABASE_URL detected in .env

PART OVERRIDES  (direct value takes priority over var name)
  -Driver   <val>   -DriverVar   <VAR>   default env var: DB_DRIVER
  -DbHost   <val>   -HostVar     <VAR>   default env var: DB_HOST
  -Port     <val>   -PortVar     <VAR>   default env var: DB_PORT
  -User     <val>   -UserVar     <VAR>   default env var: DB_USER
  -Password <val>   -PasswordVar <VAR>   default env var: DB_PASSWORD
  -DbName   <val>   -DbNameVar   <VAR>   default env var: DB_NAME
  -SslMode  <val>   -SslModeVar  <VAR>   default env var: DB_SSLMODE

EXAMPLES
  .\scripts\migrate.ps1 up
  .\scripts\migrate.ps1 down -Count 3
  .\scripts\migrate.ps1 create -MigrationName add_users_table
  .\scripts\migrate.ps1 create
  .\scripts\migrate.ps1 force -Version 5
  .\scripts\migrate.ps1 -Dir .\db\migrations up
  .\scripts\migrate.ps1 -Env .env.staging -Dir .\db\migrations up
  .\scripts\migrate.ps1 -Url "postgres://user:pass@localhost:5432/db" up
  .\scripts\migrate.ps1 -UrlVar DATABASE_URL up
"@
    exit 1
}

function Assert-Command {
    param([string]$Cmd)
    if (-not (Get-Command $Cmd -ErrorAction SilentlyContinue)) {
        Write-Err "Command '$Cmd' not found. Please install it first."
        exit 1
    }
}

function Confirm-Action {
    param([string]$Prompt)
    $r = Read-Host "$Prompt [y/N]"
    return $r.ToLower() -eq 'y'
}

# --- LOAD .env ----------------------------------------------------------------

function Import-EnvFile {
    param([string]$EnvFile)

    if (-not (Test-Path $EnvFile)) {
        Write-Err ".env file not found: $EnvFile"
        exit 1
    }

    Write-Info "Loading env from: $EnvFile"

    Get-Content $EnvFile | ForEach-Object {
        $line = $_.Trim()
        if ([string]::IsNullOrWhiteSpace($line) -or $line.StartsWith('#')) { return }

        if ($line -match '^([A-Za-z_][A-Za-z0-9_]*)=(.*)$') {
            $key   = $Matches[1].Trim()
            $value = $Matches[2].Trim()

            if    ($value -match '^"([^"]*)"')  { $value = $Matches[1] }
            elseif ($value -match "^'([^']*)'") { $value = $Matches[1] }
            else  { $value = ($value -replace '\s+#.*$', '').TrimEnd() }

            [System.Environment]::SetEnvironmentVariable($key, $value, 'Process')
        }
    }
}

# --- FIELD RESOLVER -----------------------------------------------------------
# Priority: direct value → named env var → default DB_* env var

function Resolve-Field {
    param([string]$Direct, [string]$VarName, [string]$DefaultVar)

    if ($Direct)  { return $Direct }
    if ($VarName) { return ([System.Environment]::GetEnvironmentVariable($VarName) ?? '') }
    return ([System.Environment]::GetEnvironmentVariable($DefaultVar) ?? '')
}

# --- URL-ENCODE ---------------------------------------------------------------

function ConvertTo-UrlEncoded {
    param([string]$Raw)
    if ([string]::IsNullOrEmpty($Raw)) { return '' }
    return [uri]::EscapeDataString($Raw)
}

# --- DSN BUILDER --------------------------------------------------------------

function Build-Dsn {
    $driver  = Resolve-Field $Driver   $DriverVar   'DB_DRIVER'
    $hostR   = Resolve-Field $DbHost   $HostVar     'DB_HOST'
    $portR   = Resolve-Field $Port     $PortVar     'DB_PORT'
    $userR   = Resolve-Field $User     $UserVar     'DB_USER'
    $passRaw = Resolve-Field $Password $PasswordVar 'DB_PASSWORD'
    $nameR   = Resolve-Field $DbName   $DbNameVar   'DB_NAME'
    $sslR    = Resolve-Field $SslMode  $SslModeVar  'DB_SSLMODE'
    if (-not $sslR) { $sslR = 'disable' }

    if (-not $driver) {
        Write-Err "DB driver not set. Use -Driver, -DriverVar, or DB_DRIVER in .env."
        exit 1
    }

    $pass  = ConvertTo-UrlEncoded $passRaw
    $creds = if ($pass) { "${userR}:${pass}" } else { $userR }

    if ($driver -notin @('sqlite3', 'sqlite')) {
        if (-not $userR) { Write-Err "DB user not set. Use -User / -UserVar / DB_USER."; exit 1 }
        if (-not $hostR) { Write-Err "DB host not set. Use -DbHost / -HostVar / DB_HOST."; exit 1 }
        if (-not $nameR) { Write-Err "DB name not set. Use -DbName / -DbNameVar / DB_NAME."; exit 1 }
    }

    switch ($driver) {
        { $_ -in 'postgres','postgresql','cockroachdb','redshift' } {
            $p = if ($portR) { $portR } else { '5432' }
            return "postgres://${creds}@${hostR}:${p}/${nameR}?sslmode=${sslR}"
        }
        { $_ -in 'mysql','mariadb' } {
            $p = if ($portR) { $portR } else { '3306' }
            return "mysql://${creds}@tcp(${hostR}:${p})/${nameR}"
        }
        { $_ -in 'sqlite3','sqlite' } {
            if (-not $nameR) { Write-Err "DB_NAME must be the SQLite file path."; exit 1 }
            return "sqlite3://${nameR}"
        }
        { $_ -in 'sqlserver','mssql' } {
            $p = if ($portR) { $portR } else { '1433' }
            return "sqlserver://${creds}@${hostR}:${p}?database=${nameR}"
        }
        'cassandra' {
            $p = if ($portR) { $portR } else { '9042' }
            return "cassandra://${hostR}:${p}/${nameR}"
        }
        'clickhouse' {
            $p = if ($portR) { $portR } else { '9000' }
            return "clickhouse://${hostR}:${p}/${nameR}"
        }
        default {
            Write-Err "Unknown driver '$driver'. Supported: postgres, cockroachdb, redshift,"
            Write-Err "  mysql, mariadb, sqlite3, sqlserver, mssql, cassandra, clickhouse"
            exit 1
        }
    }
}

# --- RESOLVE DSN --------------------------------------------------------------

function Resolve-Dsn {
    if ($Url) {
        Write-Info "Using -Url DSN."
        return $Url
    }
    if ($UrlVar) {
        $val = [System.Environment]::GetEnvironmentVariable($UrlVar)
        if ([string]::IsNullOrWhiteSpace($val)) {
            Write-Err "-UrlVar '$UrlVar' is empty or not found in .env."; exit 1
        }
        Write-Info "Using DSN from `$$UrlVar."
        return $val
    }
    $dbUrl = [System.Environment]::GetEnvironmentVariable('DATABASE_URL')
    if (-not [string]::IsNullOrWhiteSpace($dbUrl)) {
        Write-Info "Auto-detected DATABASE_URL."
        return $dbUrl
    }
    return Build-Dsn
}

# --- COMMANDS -----------------------------------------------------------------

function Invoke-Up {
    param([string]$Dsn, [string]$MigrationsDir)
    Write-Info "Running all pending migrations..."
    & migrate -path $MigrationsDir -database $Dsn up
    Write-Info "Done."
}

function Invoke-Down {
    param([string]$Dsn, [string]$MigrationsDir, [string]$CountArg)
    $count = if ($CountArg) { $CountArg } else { '1' }
    if ($count -notmatch '^\d+$') {
        Write-Err "Count must be a positive integer, got: '$count'"; exit 1
    }
    Write-Warn "About to roll back $count migration(s)."
    if (Confirm-Action 'Continue?') {
        & migrate -path $MigrationsDir -database $Dsn down $count
        Write-Info "Rolled back $count migration(s)."
    } else { Write-Info "Aborted." }
}

function Invoke-Create {
    param([string]$MigrationsDir, [string]$NameArg)

    # Use provided name or fall back to a timestamped default.
    # Get-Date -Format formats the current date/time into a string.
    $name = if ($NameArg) {
        $NameArg
    } else {
        "migration_$(Get-Date -Format 'yyyyMMdd_HHmmss')"
    }

    if (-not $NameArg) {
        Write-Info "No -MigrationName given. Using generated name: $name"
    }

    & migrate create -ext sql -dir $MigrationsDir -seq $name
    Write-Info "Created migration: $name"
}

function Invoke-Force {
    param([string]$Dsn, [string]$MigrationsDir, [string]$VersionArg)
    if ([string]::IsNullOrWhiteSpace($VersionArg)) {
        Write-Err "Version required. Example: .\migrate.ps1 force -Version 5"; exit 1
    }
    Write-Warn "Forcing migration version to $VersionArg."
    & migrate -path $MigrationsDir -database $Dsn force $VersionArg
    Write-Info "Version forced to $VersionArg."
}

# --- MAIN ---------------------------------------------------------------------

if (-not $Command) { Show-Usage }

Assert-Command 'migrate'

$envPath = if ($Env) { $Env } else { $DefaultEnvPath }
Import-EnvFile -EnvFile $envPath

# Resolve migrations dir; Resolve-Path -Relative handles both absolute and relative paths
$migrationsDir = if ($Dir) { $Dir } else { $DefaultMigrationsPath }

# 'create' doesn't need a DSN — run it before resolve_dsn so that DB_* vars
# being absent (fresh repo) doesn't cause a validation error.
if ($Command -eq 'create') {
    Invoke-Create -MigrationsDir $migrationsDir -NameArg $MigrationName
    exit 0
}

$dsn = Resolve-Dsn

switch ($Command) {
    'up'    { Invoke-Up    -Dsn $dsn -MigrationsDir $migrationsDir }
    'down'  { Invoke-Down  -Dsn $dsn -MigrationsDir $migrationsDir -CountArg $Count }
    'force' { Invoke-Force -Dsn $dsn -MigrationsDir $migrationsDir -VersionArg $Version }
}