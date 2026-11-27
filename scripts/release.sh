#!/bin/sh
# Build checksum-backed release archives locally, named like the release
# workflow's assets. Adapted from entireio/entire-brain (MIT, see NOTICE).
set -eu

repo_root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

version=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || printf 'dev')}
bare=${version#v}
out_dir=${OUT_DIR:-dist/release-$version}
targets=${MF_RELEASE_TARGETS:-"darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64"}

mkdir -p "$out_dir"
rm -f "$out_dir"/checksums.txt

sha256_file() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1"
	else
		shasum -a 256 "$1"
	fi
}

build_target() {
	goos=${1%/*}
	goarch=${1#*/}
	if [ "$goos" = "$goarch" ] || [ -z "$goos" ] || [ -z "$goarch" ]; then
		printf 'invalid target %s; expected GOOS/GOARCH\n' "$1" >&2
		return 1
	fi

	bin="mf-digisac"
	case "$goos" in
		windows) bin="$bin.exe" ;;
	esac

	asset="mf-digisac_${bare}_${goos}_${goarch}"
	work="$out_dir/$asset"
	rm -rf "$work"
	mkdir -p "$work"

	printf 'building %s/%s\n' "$goos" "$goarch" >&2
	GOOS="$goos" GOARCH="$goarch" CGO_ENABLED="${CGO_ENABLED:-0}" \
		go build -trimpath \
		-ldflags "-s -w -X github.com/mguilarducci/mf-digisac/internal/cli.Version=$bare" \
		-o "$work/$bin" ./cmd/mf-digisac
	cp README.md LICENSE NOTICE mf-plugin.yml "$work/"

	if [ "$goos" = windows ]; then
		(cd "$work" && zip -q "../$asset.zip" ./*)
		archive="$asset.zip"
	else
		tar -C "$work" -czf "$out_dir/$asset.tar.gz" .
		archive="$asset.tar.gz"
	fi
	rm -rf "$work"
	(cd "$out_dir" && sha256_file "$archive") >> "$out_dir/checksums.txt"
}

for target in $targets; do
	build_target "$target"
done

printf 'wrote %s\n' "$out_dir" >&2
