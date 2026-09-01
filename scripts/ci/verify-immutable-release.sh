#!/usr/bin/env bash
set -euo pipefail

# Read-only verification of a release created by an explicitly authorized
# operator. It never calls the administration-settings API and never creates,
# edits, overwrites, or deletes a release.
release_id="${1:?release id required}"
asset_dir="${2:?directory containing downloaded release assets required}"
release_json="$(gh api "repos/${GITHUB_REPOSITORY}/releases/${release_id}")"
test "$(jq -r '.immutable' <<<"$release_json")" = true

while IFS=$'\t' read -r asset_name asset_digest; do
  test -n "$asset_name"
  test -n "$asset_digest"
  local_path="${asset_dir}/${asset_name}"
  test -f "$local_path"
  local_digest="sha256:$(sha256sum "$local_path" | cut -d' ' -f1)"
  test "$local_digest" = "$asset_digest"
done < <(gh api "repos/${GITHUB_REPOSITORY}/releases/${release_id}/assets" --paginate --jq '.[] | [.name, .digest] | @tsv')
