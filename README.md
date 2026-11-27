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

Push a `vX.Y.Z` tag. `.github/workflows/release.yml` builds every target with
`scripts/release.sh`, smoke-runs the binaries, and publishes a GitHub release.
`mise run release` builds the same archives locally under `dist/`.
