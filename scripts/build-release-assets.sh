#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "Usage: $0 <tag> [output_dir]" >&2
  echo "Example: $0 v1.2.3 dist" >&2
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
  exit 1
fi

tag="$1"
output_dir="${2:-dist}"

if [[ "$tag" != v* ]]; then
  echo "tag must start with v (got: $tag)" >&2
  exit 1
fi

version="${tag#v}"
binary_name="dpskmon"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
staging_dir="$(mktemp -d)"

cleanup() {
  rm -rf "$staging_dir"
}
trap cleanup EXIT

checksum_file="$repo_root/$output_dir/checksums.txt"

checksum() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi

  shasum -a 256 "$1" | awk '{print $1}'
}

mkdir -p "$repo_root/$output_dir"
: >"$checksum_file"

targets=(
  "darwin amd64"
  "darwin arm64"
  "linux amd64"
  "linux arm64"
)

for target in "${targets[@]}"; do
  read -r goos goarch <<<"$target"

  build_dir="$staging_dir/${goos}_${goarch}"
  archive_name="${binary_name}_${version}_${goos}_${goarch}.tar.gz"
  archive_path="$repo_root/$output_dir/$archive_name"

  mkdir -p "$build_dir"

  (
    cd "$repo_root"
    env CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
      go build -trimpath -ldflags="-s -w" -o "$build_dir/$binary_name" .
  )

  tar -C "$build_dir" -czf "$archive_path" "$binary_name"
  printf "%s  %s\n" "$(checksum "$archive_path")" "$archive_name" >>"$checksum_file"
done

echo "Built release assets in $repo_root/$output_dir"
