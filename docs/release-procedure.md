# Immutable release verification

This repository has no automatic release authority. A release operator must
perform the following sequence with explicitly granted, caller-owned
credentials:

1. Create a draft release for an existing immutable tag.
2. Upload the generated source, conformance report, and evidence assets to the
   draft.
3. Publish the draft.
4. Query `GET /repos/{owner}/{repo}/releases/{release_id}` through the public
   release API and require `immutable == true`.
5. Query each uploaded asset and compare the API-provided digest with the
   locally computed SHA-256 digest.

The order is material: assets are never uploaded after publication, and a
published tag/release is never overwritten or deleted. The verification must
not query a GitHub administration-settings endpoint. The CI workflow has
`contents: read` only; it builds evidence and uploads it as a workflow
artifact, leaving release creation to an authorized human/API client.
