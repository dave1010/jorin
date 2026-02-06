# Release Process

This project ships binaries via GitHub Releases and publishes an npm package
that bundles those binaries for `npx` usage. Releases are tag-driven and fully
automated by GitHub Actions.

## Versioning

- Tags use `vX.Y.Z` (semver style).
- The Go binary embeds the tag via `-ldflags` in the release workflow.
- `package.json`/`package-lock.json` should match the next version for
  consistency, even though the npm publish workflow also sets the version
  from the tag.

## Checklist

1. Update `CHANGELOG.md`.
1. Bump npm metadata.
1. Run local checks.
1. Commit release changes.
1. Create and push a tag.
1. Publish the GitHub Release.
1. Verify npm publish.

## Detailed Steps

1. Update `CHANGELOG.md`
   - Move entries from `Unreleased` into a new version section
     like `## vX.Y.Z - YYYY-MM-DD`.
   - Leave `Unreleased` in place with `(nothing yet)`.

1. Bump npm metadata
   - Update `package.json` and `package-lock.json` to the new version.

1. Run local checks

   ```bash
   make fmt
   make test
   make lint
   ```

1. Commit release changes
   - Commit the `CHANGELOG.md` and `package*.json` updates before tagging.

1. Tag the release and push

   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

   Tag pushes trigger `.github/workflows/release.yml` which builds
   and uploads release binaries to the GitHub Release.

1. Publish the GitHub Release
   - Create the GitHub Release for the tag (or let the workflow do it if the
     tag trigger already created it).
   - Ensure the release is **published**, not a draft.

   Publishing the release triggers `.github/workflows/npm-release.yml`, which:
   - Sets the npm package version from the tag.
   - Downloads GitHub release assets into `dist/`.
   - Publishes the package with npm trusted publishing (OIDC).

1. Verify
   - Check the GitHub Release has all expected assets.
   - Confirm `npm view jorin` shows the new version.

## When Adding Platforms

If you add or remove platforms/architectures, update these in lockstep:

- `.github/workflows/release.yml` build matrix.
- `.github/workflows/npm-release.yml` asset list.
- Any docs that mention supported platforms (README or usage docs).
