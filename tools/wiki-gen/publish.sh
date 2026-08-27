#!/usr/bin/env bash
# Publish the generated wiki/ mirror to the GitHub wiki
# (github.com/eparodi/deepcut-skills/wiki, remote deepcut-skills.wiki.git).
#
# Flow: regenerate -> verify (go test) -> clone the wiki -> add/update the
# generated pages -> push. Never deletes wiki pages that are not generated;
# those are listed for manual review instead.
#
# Usage: ./tools/wiki-gen/publish.sh   (from the repo root)
set -euo pipefail

cd "$(dirname "$0")/../.."

WIKI_REMOTE="git@github.com:eparodi/deepcut-skills.wiki.git"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "==> Regenerating wiki/..."
go run ./tools/wiki-gen

echo "==> Verifying (go test ./...)..."
go test ./... -count=1

echo "==> Cloning the wiki repo..."
git clone -q "$WIKI_REMOTE" "$TMP/wiki"

echo "==> Syncing generated pages (add/update only)..."
changed=0
for f in wiki/*.md; do
	name="$(basename "$f")"
	if ! cmp -s "$f" "$TMP/wiki/$name"; then
		cp "$f" "$TMP/wiki/$name"
		echo "  updated: $name"
		changed=1
	fi
done

# List wiki pages that are not part of the generated set. Never delete them —
# the wiki may carry hand-authored pages; removing is a manual decision.
for f in "$TMP"/wiki/*.md; do
	name="$(basename "$f")"
	if [ ! -f "wiki/$name" ]; then
		echo "NOTE: '$name' exists on the wiki but is not generated — left untouched (remove manually if intended)."
	fi
done

if [ "$changed" -eq 0 ]; then
	echo "==> Wiki is already up to date; nothing to push."
	exit 0
fi

echo "==> Committing and pushing..."
git -C "$TMP/wiki" add .
git -C "$TMP/wiki" -c user.name="skills-test bot" -c user.email="noreply@example.com" commit -m "chore: update wiki (generated from skills)"
git -C "$TMP/wiki" push origin master

echo "==> Done."
