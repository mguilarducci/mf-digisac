# mf-digisac

[mf](https://github.com/mguilarducci/mf-cli) plugin for the Digisac API. Runs as `mf digisac`, or standalone as `mf-digisac`.

API reference: [Digisac API docs (Postman)](https://documenter.getpostman.com/view/24605757/2sA3BhfaDg).

## Install

```sh
# mf itself
curl -fsSL https://raw.githubusercontent.com/mguilarducci/mf-cli/main/scripts/install.sh | bash

# this plugin: downloads the release, verifies its checksum, runs `mf plugin install`
curl -fsSL https://raw.githubusercontent.com/mguilarducci/mf-digisac/main/scripts/get-digisac.sh | bash

# or, with mf already installed
mf plugin install digisac
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

1. `mise run release vX.Y.Z` tags and pushes.
2. `.github/workflows/release.yml` runs GoReleaser and publishes the GitHub
   release, with notes GitHub generates from the PRs merged since the last tag.

`mise run release:snapshot` builds the same archives locally under `dist/`.
