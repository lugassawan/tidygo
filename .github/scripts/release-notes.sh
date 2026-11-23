#!/bin/sh
# Print grouped release notes (markdown) for a git revision range.
# Usage: release-notes.sh <git-range>
#
# Commits are classified by the Conventional-Commit prefix that
# .githooks/commit-msg enforces: "feat: " and "fix: " get their own sections
# (prefix stripped); every other subject lands in "Other Changes" verbatim.
set -eu

RANGE=${1:?usage: release-notes.sh <git-range>}

SUBJECTS=$(git log "$RANGE" --no-merges --pretty=format:%s)

FEATURES=$(printf '%s\n' "$SUBJECTS" | sed -n 's/^feat: /- /p')
FIXES=$(printf '%s\n' "$SUBJECTS" | sed -n 's/^fix: /- /p')
OTHER=$(printf '%s\n' "$SUBJECTS" | sed -n -e '/^feat: /d' -e '/^fix: /d' -e '/^$/d' -e 's/^/- /p')

section() {
	if [ -n "$2" ]; then
		printf '\n### %s\n%s\n' "$1" "$2"
	fi
}

echo "## What's Changed"
section "Features" "$FEATURES"
section "Bug Fixes" "$FIXES"
section "Other Changes" "$OTHER"

if [ -z "$FEATURES$FIXES$OTHER" ]; then
	printf '\n- Maintenance release\n'
fi
