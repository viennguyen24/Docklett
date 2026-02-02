#!/bin/sh
# Setup script: Configure Git to use .githooks directory

# Ensure we're in a Git repository
if [ ! -d ".git" ]; then
    echo "Error: Not a Git repository. Run this from repo root."
    exit 1
fi

# Configure Git to use .githooks directory
echo "Configuring Git to use .githooks/ directory..."
git config core.hooksPath .githooks

# Make hooks executable
chmod +x .githooks/pre-commit

echo "✓ Git hooks configured successfully"
echo "Hook will run automatically on git commit."
echo "Note: Hook updates will be automatically synced via git pull."
