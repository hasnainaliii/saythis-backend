# Get the absolute path of the directory this script is in
$scriptDir = $PSScriptRoot

# Define the root of your project (one folder up)
$projectRoot = Join-Path -Path $scriptDir -ChildPath ".."
$envPath = Join-Path -Path $projectRoot -ChildPath ".env"
$migrationsPath = Join-Path -Path $projectRoot -ChildPath "migrations"

# 1. Load .env safely
if (Test-Path $envPath) {
    Get-Content $envPath | ForEach-Object {
        if ($_ -match '^([^#][^=]+)=(.+)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [System.Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
} else {
    Write-Host "Warning: .env file not found at $envPath" -ForegroundColor Yellow
}

# 2. Check if migrate CLI is installed
if (-not (Get-Command "migrate" -ErrorAction SilentlyContinue)) {
    Write-Error "The 'migrate' CLI tool is not installed or not in your PATH. Please install it first."
    return
}

# 3. Handle Commands
$command = $args[0]
$name = $args[1]

switch ($command) {
    "up" {
        migrate -path $migrationsPath -database $env:DATABASE_URL up
    }

    "down" {
        $count = if ($name) { $name } else { "1" }
        $confirm = Read-Host "Rolling back $count migration(s). Continue? [y/N]"
        if ($confirm -eq 'y') {
            migrate -path $migrationsPath -database $env:DATABASE_URL down $count
        }
    }

    "create" {
        if (-not $name) { Write-Error "Migration name is required"; return }
        migrate create -ext sql -dir $migrationsPath -seq $name
    }

    "force" {
        if (-not $name) { Write-Error "Version number is required"; return }
        migrate -path $migrationsPath -database $env:DATABASE_URL force $name
    }

    Default {
        Write-Host "Usage: .\migrate.ps1 [up|down|create|force] [args]"
    }
}
