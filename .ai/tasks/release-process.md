# snag - Release Process

> **Purpose**: Repeatable process for releasing new versions of snag
> **Audience**: AI agents and maintainers performing releases
> **Last Updated**: 2025-10-29

This document provides step-by-step instructions for releasing snag. Execute each step in order.

---

## Prerequisites

Verify before starting:

- Write access to `grantcarthew/snag` repository
- Write access to `grantcarthew/homebrew-tap` repository
- Go 1.25.3+ installed
- Git configured with proper credentials
- GitHub CLI (`gh`) installed and authenticated
- All planned features/fixes merged to main branch

---

## Release Process

**Steps**:

1. Run pre-release validation
2. Determine version number
3. Update CHANGELOG.md
4. Commit changes
5. Create and push git tag
6. Create GitHub Release
7. Update Homebrew tap
8. Verify installation
9. Clean up

**Estimated Time**: 20-30 minutes

---

## Step 1: Pre-Release Validation

Run validation checks:

```bash
# Ensure on main branch with latest changes
git checkout main
git pull origin main

# Skip test suite for this release? (if recently verified)
# Verify all tests pass
go test -v ./...

# Verify build works
go build -o snag ./cmd/snag
./snag --version
rm snag

# Verify clean working directory
git status
```

**Expected results**:

- All tests pass
- Build completes without errors
- `git status` shows clean working tree
- Documentation is current

**If any validation fails, stop and fix issues before proceeding.**

---

## Step 2: Determine Version Number

Set the version number using [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking API changes (1.0.0 → 2.0.0)
- **MINOR**: New features, backward compatible (1.0.0 → 1.1.0)
- **PATCH**: Bug fixes only (1.0.0 → 1.0.1)

```bash
# Check current version
git tag -l | tail -1

# Set new version (example: v0.1.0)
export VERSION="0.1.0"
echo "Releasing version: v${VERSION}"
```

---

## Step 3: Update CHANGELOG.md

Review changes since last release and update CHANGELOG.md:

```bash
# Show changes since previous version
PREV_VERSION=$(git tag -l | tail -1)
echo "Changes since ${PREV_VERSION}:"
git log ${PREV_VERSION}..HEAD --oneline

# Review the changes and categorize them
# Then edit CHANGELOG.md manually
```

Update CHANGELOG.md by adding a new version section with:

- **Added**: New features
- **Changed**: Changes to existing functionality
- **Fixed**: Bug fixes
- **Deprecated**: Features marked for removal
- **Removed**: Removed features
- **Security**: Security fixes

Example format:

```markdown
## [0.1.0] - 2025-10-29

### Added

- New feature X for better Y

### Fixed

- Bug in component A

[Unreleased]: https://github.com/grantcarthew/snag/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/grantcarthew/snag/releases/tag/v0.1.0
```

---

## Step 4: Commit Changes

Commit the CHANGELOG:

```bash
# Stage and commit changes
git add CHANGELOG.md
git commit -m "chore: prepare for v${VERSION} release"
git push origin main
```

---

## Step 5: Create and Push Git Tag

Create an annotated git tag:

```bash
# Get previous version and review changes
PREV_VERSION=$(git tag -l | tail -1)
git log ${PREV_VERSION}..HEAD --oneline

# Create one-line summary from the changes above
# Examples: "Add tab management features", "Fix authentication handling"
SUMMARY="Your one-line summary here"

# Create and push annotated tag
git tag -a "v${VERSION}" -m "Release v${VERSION} - ${SUMMARY}"
git push origin "v${VERSION}"

# Verify tag exists
git tag -l -n9 "v${VERSION}"
```

---

## Step 6: Create GitHub Release

Create the GitHub Release with release notes:

```bash
# Wait for tarball to be generated (usually immediate)
sleep 5

# Get tarball SHA256 for Homebrew (will use in Step 7)
TARBALL_URL="https://github.com/grantcarthew/snag/archive/refs/tags/v${VERSION}.tar.gz"
# Linux:
TARBALL_SHA256=$(curl -sL "$TARBALL_URL" | sha256sum | cut -d' ' -f1)
# macOS:
TARBALL_SHA256=$(curl -sL "$TARBALL_URL" | shasum -a 256 | cut -d' ' -f1)
echo "Tarball SHA256: $TARBALL_SHA256"

# Create GitHub Release using gh CLI
# Extract release notes from CHANGELOG.md for this version
# Or create release notes based on git log

gh release create "v${VERSION}" \
  --title "Release v${VERSION}" \
  --notes "$(cat <<EOF
## Changes

$(git log ${PREV_VERSION}..v${VERSION} --pretty=format:"- %s" --reverse)

See [CHANGELOG.md](https://github.com/grantcarthew/snag/blob/main/CHANGELOG.md) for details.
EOF
)"

# Verify release was created
gh release view "v${VERSION}"
```

**Note**: GitHub automatically attaches source archives (tar.gz, zip) to releases. Homebrew builds from the tar.gz archive.

---

## Step 7: Update Homebrew Tap

Update the Homebrew formula with the new version:

```bash
# Navigate to homebrew-tap directory
cd ~/Projects/homebrew-tap
git pull origin main

# Display tarball info from Step 6
echo "Tarball URL: $TARBALL_URL"
echo "Tarball SHA256: $TARBALL_SHA256"

# Edit Formula/snag.rb and update:
# 1. url line: Update version in URL
# 2. sha256 line: Update with TARBALL_SHA256
# 3. ldflags: Update version in "-X github.com/p3bot/snag/internal/cli.Version=X.X.X"
# 4. test: Update expected version in assert_match

# After editing, commit and push
git add Formula/snag.rb
git commit -m "snag: update to ${VERSION}"
git push origin main

# Return to snag directory
cd -
```

**Formula example** (Formula/snag.rb):

```ruby
url "https://github.com/grantcarthew/snag/archive/refs/tags/v0.1.0.tar.gz"
sha256 "abc123..."  # Use TARBALL_SHA256 value
ldflags = "-X github.com/p3bot/snag/internal/cli.Version=0.1.0"
system "go", "build", *std_go_args(ldflags:), "./cmd/snag"
assert_match "0.1.0", shell_output("#{bin}/snag --version")
```

---

## Step 8: Verify Installation

Test the Homebrew installation:

```bash
# Update and reinstall
brew update
brew reinstall grantcarthew/tap/snag

# Verify version
snag --version  # Should show new version

# Test basic functionality
snag --quiet https://example.com
```

**Expected results**:

- `snag --version` displays new version
- Basic page fetch returns markdown content
- No errors during installation

**If installation fails**, debug with:

```bash
brew audit --strict grantcarthew/tap/snag
brew install --verbose grantcarthew/tap/snag
```

---

## Step 9: Clean Up

Complete the release:

```bash
# Verify release is live
gh release view "v${VERSION}"

# Check Homebrew tap was updated
cd ~/Projects/homebrew-tap
git log -1
cd -

# Clean up any artifacts
rm -rf dist/

# Verify clean state
git status
```

**Release is complete!**

Monitor for issues:

- Watch GitHub issues for bug reports
- Monitor Homebrew installation feedback
- Be ready to release a patch if critical issues arise

---

## Rollback Procedure

If critical issues are discovered after release:

**Option 1: Patch Release** (Recommended)

```bash
# Fix the issue, then release patch version (e.g., v0.1.1)
# Follow the standard release process
```

**Option 2: Delete Release** (Last resort - use only for critical security issues)

```bash
# Delete GitHub release
gh release delete "v${VERSION}" --yes

# Delete tags
git push origin --delete "v${VERSION}"
git tag -d "v${VERSION}"

# Revert Homebrew tap
cd ~/Projects/homebrew-tap
git revert HEAD
git push origin main
cd -
```

---

## Quick Reference

One-command release workflow:

```bash
# Set version
export VERSION="0.1.0"

# Get previous version for change summary
PREV_VERSION=$(git tag -l | tail -1)

# 1. Validation
go test -v ./...
git status  # Should be clean

# 2. Update CHANGELOG.md manually, then commit
git add CHANGELOG.md
git commit -m "chore: prepare for v${VERSION} release"
git push origin main

# 3. Create tag with summary
git log ${PREV_VERSION}..HEAD --oneline  # Review changes
SUMMARY="Your summary here"
git tag -a "v${VERSION}" -m "Release v${VERSION} - ${SUMMARY}"
git push origin "v${VERSION}"

# 4. Create GitHub Release
gh release create "v${VERSION}" --title "Release v${VERSION}" \
  --notes "$(git log ${PREV_VERSION}..v${VERSION} --pretty=format:'- %s')"

# 5. Get tarball SHA256
# Linux: sha256sum | macOS: shasum -a 256
TARBALL_SHA256=$(curl -sL "https://github.com/grantcarthew/snag/archive/refs/tags/v${VERSION}.tar.gz" | shasum -a 256 | cut -d' ' -f1)
echo "SHA256: $TARBALL_SHA256"

# 6. Update Homebrew (edit Formula/snag.rb with VERSION and SHA256)
cd ~/Projects/homebrew-tap
# Edit Formula/snag.rb
git add Formula/snag.rb
git commit -m "snag: update to ${VERSION}"
git push origin main
cd -

# 7. Test
brew update && brew reinstall grantcarthew/tap/snag
snag --version
```

---

## Troubleshooting

**Tests failing**

- Run: `go test -v ./...` to see detailed output
- Fix all failures before proceeding
- Never release with failing tests

**Tarball not available**

- Wait 1-2 minutes after pushing tag
- Verify tag exists: `git ls-remote --tags origin | grep v${VERSION}`
- Check: https://github.com/grantcarthew/snag/tags

**Homebrew formula issues**

- Audit: `brew audit --strict grantcarthew/tap/snag`
- Common: Incorrect SHA256, wrong URL format, Ruby syntax
- Fix and push updated formula

**Installation fails**

- Verbose output: `brew install --verbose grantcarthew/tap/snag`
- View formula: `brew cat grantcarthew/tap/snag`
- Verify tarball: `curl -I https://github.com/grantcarthew/snag/archive/refs/tags/v${VERSION}.tar.gz`

---

## Related Documents

- `AGENTS.md` - Repository context for AI agents
- `CHANGELOG.md` - Version history
- `.ai/design/design-records/dr-001-snag-design.md` - Design decisions and rationale
- `README.md` - User-facing documentation

---

**End of Release Process**
