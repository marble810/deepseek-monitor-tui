#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "Usage: $0 <tag> [artifacts_dir]" >&2
  echo "Example: $0 v1.2.3 dist" >&2
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required" >&2
  exit 1
fi

if [[ -z "${GITHUB_TOKEN:-}" && -z "${GH_TOKEN:-}" ]]; then
  echo "GITHUB_TOKEN or GH_TOKEN is required" >&2
  exit 1
fi

tag="$1"
artifacts_dir="${2:-dist}"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

shopt -s nullglob
artifacts=("$repo_root/$artifacts_dir"/*)
shopt -u nullglob

if [[ ${#artifacts[@]} -eq 0 ]]; then
  echo "no artifacts found in $repo_root/$artifacts_dir" >&2
  exit 1
fi

if gh release view "$tag" >/dev/null 2>&1; then
  gh release upload "$tag" "${artifacts[@]}" --clobber
  exit 0
fi

gh release create "$tag" "${artifacts[@]}" \
  --title "$tag" \
  --verify-tag \
  --generate-notes
