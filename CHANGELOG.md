# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-11-27

### Added

- `mf digisac contacts list` lists Digisac contacts as a table or with `-o json`, one page or every page with `--all` ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))
- `mf digisac auth login|status|logout` saves, checks and removes the Digisac URL and API token. The token is read without echo or from stdin, never from a flag, and `DIGISAC_URL` / `DIGISAC_TOKEN` override the saved values ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))
- `mf digisac version` prints the plugin version ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))
- `scripts/get-digisac.sh` installs the latest release, verified against the release checksums, and registers it with `mf plugin install` ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))

### Housekeeping

- Layout, mise tasks, lint and CI follow entireio/entire-brain, and releases follow mf-cli ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))
