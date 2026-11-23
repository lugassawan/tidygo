#!/bin/sh
# Print grouped release notes (markdown) for a git revision range.
# Usage: release-notes.sh <git-range>
#
# Commits are classified by the Conventional-Commit prefix that
# .githooks/commit-msg enforces: "feat: " and "fix: " get their own sections;
# every other subject lands in "Other Changes". Every section has its
# "type: " prefix stripped and its bullets capitalized.
set -eu

RANGE=${1:?usage: release-notes.sh <git-range>}

SUBJECTS=$(git log "$RANGE" --no-merges --pretty=format:%s)

# Turn subjects on stdin into capitalized markdown bullets. awk, not sed:
# BSD sed (macOS) has no \u, so it cannot capitalize portably. LC_ALL=C keeps
# BSD awk from aborting when a subject starts with a multibyte character.
bullets() {
	LC_ALL=C awk 'NF { print "- " toupper(substr($0, 1, 1)) substr($0, 2) }'
}

FEATURES=$(printf '%s\n' "$SUBJECTS" | sed -n 's/^feat: //p' | bullets)
FIXES=$(printf '%s\n' "$SUBJECTS" | sed -n 's/^fix: //p' | bullets)
OTHER=$(printf '%s\n' "$SUBJECTS" | sed -e '/^feat: /d' -e '/^fix: /d' -e 's/^[a-z][a-z]*: //' | bullets)

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
