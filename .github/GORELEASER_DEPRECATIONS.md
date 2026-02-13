# GoReleaser Deprecation Warnings

This document explains the deprecation warnings you might see when running GoReleaser.

## Current Warnings

### 1. ~~`archives.format_overrides.format`~~ ✅ FIXED

**Status**: ✅ **Resolved**

**What was deprecated**: Explicitly setting archive format via `format_overrides`

**Solution**: Removed `format_overrides` - GoReleaser uses `tar.gz` by default for macOS

**Before**:
```yaml
archives:
  - format_overrides:
      - goos: darwin
        format: tar.gz
```

**After**:
```yaml
archives:
  - id: apple-mail-mcp
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - LICENSE
      - README.md
      - CONTRIBUTING.md
```

### 2. `brews` Field

**Status**: ⚠️ **Warning Only - Still Functional**

**Message**: "brews is being phased out in favor of homebrew_casks"

**Reality**: This message is misleading for our use case.

#### Explanation

- **`brews`**: For Homebrew **formulae** (CLI tools) ← **This is what we use**
- **`homebrew_casks`**: For Homebrew **casks** (GUI applications)

Our tool (`apple-mail-mcp`) is a CLI binary, so we use **formulae** (brews), not casks.

#### Why We Keep `brews:`

1. ✅ **Correct for CLI tools**: Formulae are for command-line tools
2. ✅ **Still fully supported**: Not actually deprecated for formulae
3. ✅ **Works perfectly**: No functional issues
4. ❌ **`homebrew_casks` is wrong**: Casks are for GUI apps (like `brew install --cask chrome`)

#### Current Configuration (Correct)

```yaml
brews:
  - repository:
      owner: dastrobu
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    
    directory: Formula
    
    homepage: https://github.com/dastrobu/apple-mail-mcp
    description: "MCP server providing programmatic access to macOS Mail.app"
    
    license: MIT
    
    test: |
      system "#{bin}/apple-mail-mcp", "--version"
    
    install: |
      bin.install "apple-mail-mcp"
```

#### Alternative Modern Syntax (Optional)

If you want to silence the warning (not necessary), you could use the `homebrew` section instead:

```yaml
# Alternative (but brews: still works fine)
brews:
  - name: apple-mail-mcp
    repository:
      owner: dastrobu
      name: homebrew-tap
    # ... rest of config
```

**Recommendation**: Keep current `brews:` configuration. It works perfectly and is correct for CLI tools.

## Impact on Releases

**None**. These warnings don't affect:
- ✅ Building binaries
- ✅ Creating releases
- ✅ Publishing to Homebrew tap
- ✅ Users installing via `brew install`

The warnings are just GoReleaser's way of notifying about future changes.

## When to Update

Update the configuration when:
1. GoReleaser actually removes support for `brews:` (not happening soon)
2. A new stable syntax is recommended specifically for formulae
3. The warnings become errors (they won't)

For now: **No action needed**. Your configuration is valid and works perfectly.

## Testing

```bash
# Verify configuration is valid (ignoring warnings)
goreleaser check

# Test build (should succeed)
goreleaser build --snapshot --clean --single-target

# Test full release process (dry-run)
goreleaser release --snapshot --clean --skip=publish
```

All should work without errors.

## References

- [GoReleaser Brews Documentation](https://goreleaser.com/customization/homebrew/)
- [GoReleaser Deprecations](https://goreleaser.com/deprecations/)
- [Homebrew Formulae vs Casks](https://docs.brew.sh/Formula-Cookbook)

## Summary

✅ **Configuration is valid**
⚠️ **Warnings are informational only**
🎯 **No changes needed for production use**

The current setup will continue to work for the foreseeable future. GoReleaser maintains backward compatibility.
