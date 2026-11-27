# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-11-27

### Added

- `mf <name>` runs an `mf-<name>` executable from PATH, kubectl-style, passing through arguments, stdio and the exit code. Plugins get an allowlisted environment plus variables prefixed with their own name (`DIGISAC_*` for `mf-digisac`), and `MF_PLUGIN_ENV` passes extra ones ([#2](https://github.com/mguilarducci/mf-cli/pull/2))
- `mf plugin install|list|remove` manages plugins in `~/.local/share/mf/plugins`. Installs are symlinks, so rebuilding a plugin updates it, and installed plugins are listed under "Plugin Commands:" in `mf --help` ([#2](https://github.com/mguilarducci/mf-cli/pull/2))
- `scripts/install.sh` installs the latest release into `~/.local/bin`, verified against the release checksums ([#3](https://github.com/mguilarducci/mf-cli/pull/3))

### Fixed

- Plugins are found and installed on Windows, by `PATHEXT` extension instead of the executable bit ([#3](https://github.com/mguilarducci/mf-cli/pull/3))

### Housekeeping

- The Digisac plugin moved to its own repository, [mf-digisac](https://github.com/mguilarducci/mf-digisac) ([#1](https://github.com/mguilarducci/mf-cli/pull/1))
- Layout, mise tasks, lint, CI and GoReleaser release pipeline follow entireio/cli ([#3](https://github.com/mguilarducci/mf-cli/pull/3))

## [0.1.0] - 2026-11-27

### Added

- `mf digisac contacts list` lists Digisac contacts as a table or with `-o json`, one page or every page with `--all` ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))
- `mf digisac auth login|status|logout` saves, checks and removes the Digisac URL and API token. The token is read without echo or from stdin, never from a flag, and `DIGISAC_URL` / `DIGISAC_TOKEN` override the saved values ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))
- `mf digisac version` prints the plugin version ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))
- `scripts/get-digisac.sh` installs the latest release, verified against the release checksums, and registers it with `mf plugin install` ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))

### Housekeeping

- Layout, mise tasks, lint and CI follow entireio/entire-brain, and releases follow mf-cli ([#1](https://github.com/mguilarducci/mf-digisac/pull/1))
