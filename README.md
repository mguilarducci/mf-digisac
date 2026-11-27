# mf-digisac

[mf](https://github.com/mguilarducci/mf-cli) plugin for the Digisac API. Runs as `mf digisac`, or standalone as `mf-digisac`.

## Install

```sh
# mf itself
curl -fsSL https://raw.githubusercontent.com/mguilarducci/mf-cli/main/scripts/install.sh | bash

# this plugin: downloads the release, verifies its checksum, runs `mf plugin install`
curl -fsSL https://raw.githubusercontent.com/mguilarducci/mf-digisac/main/scripts/get-digisac.sh | bash
```

From a checkout: `mise run install`.

## Use

```sh
mf digisac auth login
mf digisac contacts list
```

Credentials come from `mf digisac auth login` or the `DIGISAC_URL` / `DIGISAC_TOKEN` env vars.

## Release

Same flow as [mf-cli](https://github.com/mguilarducci/mf-cli):

1. Ask Claude for "changelog for X.Y.Z". The `changelog` skill writes the
   `CHANGELOG.md` entry from the PRs merged since the last tag. Commit it as
   `X.Y.Z` on a `changelog-X-Y-Z` branch and merge it.
2. `mise run release vX.Y.Z` tags and pushes.
3. `.github/workflows/release.yml` runs GoReleaser and publishes the GitHub
   release with that `CHANGELOG.md` section as the notes.

`mise run release:snapshot` builds the same archives locally under `dist/`.
