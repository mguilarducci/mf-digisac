#!/bin/sh
# Build this checkout and register it with the local mf CLI.
# Adapted from entireio/entire-brain (MIT, see NOTICE).
set -eu

repo_root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

if ! command -v mf >/dev/null 2>&1; then
	printf 'mf CLI is required for plugin installation:\n  curl -fsSL https://raw.githubusercontent.com/mguilarducci/mf-cli/main/scripts/install.sh | bash\n' >&2
	exit 1
fi

version=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || printf 'dev')}

mkdir -p bin
CGO_ENABLED="${CGO_ENABLED:-0}" go build -trimpath \
	-ldflags "-X github.com/mguilarducci/mf-digisac/internal/cli.Version=$version" \
	-o bin/mf-digisac ./cmd/mf-digisac

mf plugin install ./bin/mf-digisac
mf digisac version
