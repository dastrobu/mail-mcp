# Release Process

This document describes how to create a new release of the Apple Mail MCP Server.

## Prerequisites

1. **Homebrew Tap Repository**: Create a GitHub repository named `homebrew-tap` under your account
   - Example: `https://github.com/dastrobu/homebrew-tap`
   - This will be used to publish Homebrew formulae

2. **GitHub Token for Homebrew**: Create a GitHub Personal Access Token with `repo` scope
   - Go to: https://github.com/settings/tokens/new
   - Select scopes: `repo` (full control of private repositories)
   - Name it something like "GoReleaser Homebrew Tap"
   - Copy the token

3. **Add Secret to Repository**:
   - Go to your repository settings: `https://github.com/dastrobu/apple-mail-mcp/settings/secrets/actions`
   - Click "New repository secret"
   - Name: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Value: Paste your GitHub token
   - Click "Add secret"

## Release Steps

### 1. Prepare for Release

Ensure all changes are committed and pushed to `main`:

```bash
git checkout main
git pull origin main
```

Verify CI is passing:
- Check: https://github.com/dastrobu/apple-mail-mcp/actions

### 2. Test GoReleaser Locally (Optional but Recommended)

Install GoReleaser:

```bash
brew install goreleaser
```

Run a dry-run build:

```bash
goreleaser release --snapshot --clean --skip=publish
```

This will:
- Build binaries for both architectures (amd64 and arm64)
- Create archives
- Generate checksums
- But NOT publish anything

Check the `dist/` directory to verify the build artifacts.

### 3. Create and Push a Version Tag

Choose a semantic version number (e.g., `v0.1.0`, `v1.0.0`, `v1.2.3`).

```bash
# Create an annotated tag
git tag -a v0.1.0 -m "Release v0.1.0"

# Push the tag to GitHub
git push origin v0.1.0
```

**Important**: Use an annotated tag (`-a` flag), not a lightweight tag.

### 4. GitHub Actions Takes Over

Once you push the tag:

1. GitHub Actions automatically triggers the release workflow
2. Watch the progress: https://github.com/dastrobu/apple-mail-mcp/actions
3. The workflow will:
   - Build binaries for macOS (amd64 and arm64)
   - Run tests
   - Create GitHub release with changelog
   - Upload release artifacts (binaries, archives, checksums)
   - Push Homebrew formula to your tap repository

### 5. Verify the Release

#### Check GitHub Release

1. Go to: https://github.com/dastrobu/apple-mail-mcp/releases
2. Verify the new release is published
3. Check that all artifacts are present:
   - `apple-mail-mcp_v0.1.0_darwin_amd64.tar.gz`
   - `apple-mail-mcp_v0.1.0_darwin_arm64.tar.gz`
   - `apple-mail-mcp_v0.1.0_checksums.txt`

#### Check Homebrew Formula

1. Go to: https://github.com/dastrobu/homebrew-tap
2. Verify the formula was created/updated in `Formula/apple-mail-mcp.rb`

#### Test Homebrew Installation

```bash
# Remove old version if installed
brew uninstall apple-mail-mcp

# Install from tap
brew tap dastrobu/tap
brew install apple-mail-mcp

# Verify installation
apple-mail-mcp --version
```

### 6. Announce the Release (Optional)

- Update the README if needed
- Post release notes on social media, forums, etc.
- Close related issues/PRs

## Troubleshooting

### Release workflow fails

1. Check the workflow logs: https://github.com/dastrobu/apple-mail-mcp/actions
2. Common issues:
   - Missing `HOMEBREW_TAP_GITHUB_TOKEN` secret
   - Tests failing
   - Invalid `.goreleaser.yaml` configuration

### Homebrew formula not updated

1. Verify the `HOMEBREW_TAP_GITHUB_TOKEN` secret is set correctly
2. Check that the token has `repo` scope
3. Verify the `homebrew-tap` repository exists and is accessible

### Want to redo a release

If you need to recreate a release:

```bash
# Delete the tag locally
git tag -d v0.1.0

# Delete the tag on GitHub
git push origin :refs/tags/v0.1.0

# Delete the GitHub release manually (via web interface)

# Create the tag again
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v1.0.0 → v2.0.0): Incompatible API changes
- **MINOR** version (v1.0.0 → v1.1.0): New features (backwards compatible)
- **PATCH** version (v1.0.0 → v1.0.1): Bug fixes (backwards compatible)

## Commit Messages

For clear changelogs, consider using conventional commits (though all non-merge commits will be included):

- `feat:` - New features
- `fix:` - Bug fixes
- `perf:` - Performance improvements
- `docs:` - Documentation changes
- `test:` - Test changes
- `ci:` - CI changes
- `chore:` - Maintenance

Example:
```
feat: add create_reply_draft tool
fix: handle Mail.app not running gracefully
perf: optimize message retrieval using whose() filter
```

## Manual Release (Fallback)

If GitHub Actions is not available, you can release manually:

```bash
# Set GitHub token
export GITHUB_TOKEN="your-github-token"
export HOMEBREW_TAP_GITHUB_TOKEN="your-homebrew-token"

# Create release
goreleaser release --clean
```

This requires GoReleaser installed locally (`brew install goreleaser`).
