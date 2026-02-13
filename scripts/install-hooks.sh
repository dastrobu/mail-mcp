#!/bin/bash
#
# Install git hooks for this repository
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GIT_DIR="$(git rev-parse --git-dir)"
HOOKS_DIR="$GIT_DIR/hooks"

echo "Installing git hooks..."

# Install pre-commit hook
if [ -f "$SCRIPT_DIR/pre-commit" ]; then
    cp "$SCRIPT_DIR/pre-commit" "$HOOKS_DIR/pre-commit"
    chmod +x "$HOOKS_DIR/pre-commit"
    echo "✓ Installed pre-commit hook"
else
    echo "✗ pre-commit hook not found at $SCRIPT_DIR/pre-commit"
    exit 1
fi

echo ""
echo "Git hooks installed successfully!"
echo "The pre-commit hook will automatically:"
echo "  - Run 'go fmt' on staged Go files"
echo "  - Update README.md Table of Contents with doctoc when README.md is staged"