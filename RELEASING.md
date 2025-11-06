# Releasing

## Create a Release

```bash
# 1. Run CI checks locally
just ci

# 2. Ensure all changes are committed and pushed
git add -A
git commit -m "chore: prepare release"
git push origin main

# 3. Create and push tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# 4. GitHub Action automatically:
#    - Builds binaries for all platforms
#    - Creates GitHub Release
#    - Uploads binaries and checksums
#    - Tracks build in Versioner
```

## Verify Release

1. Check GitHub Actions: https://github.com/versioner-io/versioner-cli/actions
2. Verify release: https://github.com/versioner-io/versioner-cli/releases
3. Check Versioner dashboard: https://app.versioner.io (filter by `versioner-cli`)

## Version Format

- **Beta releases:** `v0.x.x` (marked as pre-release)
- **Stable releases:** `v1.x.x` and above
- Follow [semantic versioning](https://semver.org/)
