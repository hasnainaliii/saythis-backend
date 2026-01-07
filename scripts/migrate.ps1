# Parse the .env file and set environment variables
Get-Content .env | ForEach-Object {
    if ($_ -match '^([^#][^=]+)=(.+)$') {
        $name = $matches[1].Trim()
        $value = $matches[2].Trim()
        Set-Item -Path "env:$name" -Value $value
    }
}

$command = $args[0]
$name = $args[1]

switch ($command) {
    "up" { 
        migrate -path migrations -database $env:DATABASE_URL up 
    }
    "down" {
        $count = if ($name) { $name } else { "1" }
        $confirm = Read-Host "Rolling back $count migration(s). Continue? [y/N]"
        if ($confirm -eq 'y') {
            migrate -path migrations -database $env:DATABASE_URL down $count
        }
    }
    "create" { 
        if (-not $name) { Write-Error "Migration name is required for 'create'"; return }
        migrate create -ext sql -dir migrations -seq $name 
    }
    "force" {
        if (-not $name) { Write-Error "Version number is required for 'force'"; return }
        migrate -path migrations -database $env:DATABASE_URL force $name
    }
    Default {
        Write-Host "Usage: ./script.ps1 [up|down|create|force] [name/count/version]"
    }
}