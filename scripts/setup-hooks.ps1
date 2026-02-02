# Setup script: Configure Git to use .githooks directory (PowerShell)

# Ensure we're in a Git repository
if (-Not (Test-Path ".git")) {
    Write-Error "Not a Git repository. Run this from repo root."
    exit 1
}

# Configure Git to use .githooks directory
Write-Host "Configuring Git to use .githooks/ directory..."
git config core.hooksPath .githooks

# Set execution permissions for Git Bash layer on Windows
Write-Host "Setting execution permissions..."
icacls .githooks\pre-commit /grant Everyone:RX | Out-Null

Write-Host "✓ Git hooks configured successfully" -ForegroundColor Green
Write-Host "Hook will run automatically on git commit"
Write-Host "Note: Hook updates will be automatically synced via git pull"
