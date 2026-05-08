#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "Usage: $0 <tag> <tap_dir>" >&2
  echo "Example: $0 v1.2.3 ./homebrew-tap" >&2
}

if [[ $# -ne 2 ]]; then
  usage
  exit 1
fi

tag="$1"
tap_dir="$2"

if [[ "$tag" != v* ]]; then
  echo "tag must start with v (got: $tag)" >&2
  exit 1
fi

version="${tag#v}"
formula_name="dpskmon"
source_repo="${SOURCE_REPOSITORY:-marble810/deepseek-monitor-tui}"
asset_dir="${ASSET_DIR:-}"
release_base_url="https://github.com/${source_repo}/releases/download/${tag}"
work_dir="$(mktemp -d)"

cleanup() {
  rm -rf "$work_dir"
}
trap cleanup EXIT

checksum() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi

  shasum -a 256 "$1" | awk '{print $1}'
}

resolve_asset() {
  local asset_name="$1"
  local asset_path

  if [[ -n "$asset_dir" ]]; then
    asset_path="$asset_dir/$asset_name"
    if [[ ! -f "$asset_path" ]]; then
      echo "missing asset in ASSET_DIR: $asset_path" >&2
      exit 1
    fi

    printf "%s\n" "$asset_path"
    return
  fi

  if ! command -v gh >/dev/null 2>&1; then
    echo "gh CLI is required when ASSET_DIR is not set" >&2
    exit 1
  fi

  gh release download "$tag" \
    --repo "$source_repo" \
    --pattern "$asset_name" \
    --dir "$work_dir" \
    --clobber >/dev/null

  printf "%s\n" "$work_dir/$asset_name"
}

darwin_amd64_asset="${formula_name}_${version}_darwin_amd64.tar.gz"
darwin_arm64_asset="${formula_name}_${version}_darwin_arm64.tar.gz"
linux_amd64_asset="${formula_name}_${version}_linux_amd64.tar.gz"
linux_arm64_asset="${formula_name}_${version}_linux_arm64.tar.gz"

darwin_amd64_sha="$(checksum "$(resolve_asset "$darwin_amd64_asset")")"
darwin_arm64_sha="$(checksum "$(resolve_asset "$darwin_arm64_asset")")"
linux_amd64_sha="$(checksum "$(resolve_asset "$linux_amd64_asset")")"
linux_arm64_sha="$(checksum "$(resolve_asset "$linux_arm64_asset")")"

mkdir -p "$tap_dir/Formula"

cat >"$tap_dir/Formula/${formula_name}.rb" <<EOF
class Dpskmon < Formula
  desc "Terminal UI for monitoring DeepSeek usage and balance"
  homepage "https://github.com/${source_repo}"
  version "${version}"

  on_macos do
    if Hardware::CPU.arm?
      url "${release_base_url}/${darwin_arm64_asset}"
      sha256 "${darwin_arm64_sha}"
    else
      url "${release_base_url}/${darwin_amd64_asset}"
      sha256 "${darwin_amd64_sha}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "${release_base_url}/${linux_arm64_asset}"
      sha256 "${linux_arm64_sha}"
    else
      url "${release_base_url}/${linux_amd64_asset}"
      sha256 "${linux_amd64_sha}"
    end
  end

  def install
    bin.install "dpskmon"
  end

  test do
    output = shell_output("#{bin}/dpskmon 2>&1", 1)
    assert_match "no platform bearer token found", output
  end
end
EOF

echo "Updated $tap_dir/Formula/${formula_name}.rb"
