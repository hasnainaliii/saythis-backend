# -----------------------------
# migrate.ps1 (Run from inside /scripts folder)
# -----------------------------

# Load .env from PROJECT ROOT (../.env)
if (Test-Path ..\.env) {
    Get-Content ..\.env | ForEach-Object {
        if ($_ -match '^([^#][^=]+)=(.+)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            Set-Item -Path "env:$name" -Value $value
        }
    }
}
else {
    Write-Warning "⚠️  .env file not found at ..\.env. Ensure you run this from the /scripts directory."
}

$command = $args[0]
$name = $args[1]

# Migrations folder relative to /scripts execution (Parent Directory)
# Use ../migrations format which works for golang-migrate on Windows
$migrationsPath = "../migrations"

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

    "drop" {
        $confirm = Read-Host "THIS WILL DELETE ALL TABLES. Are you absolutely sure? [y/N]"
        if ($confirm -eq 'y') {
            migrate -path $migrationsPath -database $env:DATABASE_URL drop -f
        }
    }

    "create" { 
        if (-not $name) { Write-Error "Migration name is required for 'create'"; return }
        migrate create -ext sql -dir $migrationsPath -seq $name 
    }

    "force" {
        if (-not $name) { Write-Error "Version number is required for 'force'"; return }
        migrate -path $migrationsPath -database $env:DATABASE_URL force $name
    }

    Default {
        Write-Host "Usage: .\scripts\migrate.ps1 [up|down|drop|create|force] [name/count/version]"
        Write-Host "NOTE: Run this script from the /scripts directory!"
    }
}
