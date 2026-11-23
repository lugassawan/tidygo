#!/bin/sh
# Regression test for release-notes.sh.
# Usage: sh .github/scripts/release-notes_test.sh
set -eu

SCRIPT="$(cd "$(dirname "$0")" && pwd)/release-notes.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

FAILED=0

pass() { echo "PASS $1"; }
fail() {
	echo "FAIL $1"
	FAILED=1
}

# Keep the throwaway repo hermetic: ignore a surrounding git hook's repo/identity
# env and the caller's global/system config (GIT_CONFIG_GLOBAL needs git >= 2.32).
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE \
	GIT_AUTHOR_NAME GIT_AUTHOR_EMAIL GIT_COMMITTER_NAME GIT_COMMITTER_EMAIL
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1

# Fixed identity, no signing, no commit hooks.
commit() {
	git -c user.name=test -c user.email=test@example.invalid -c commit.gpgsign=false \
		commit -q --allow-empty --no-verify -m "$1"
}

cd "$TMP"
git init -q

commit "initial import"
# Non-ASCII first letter stays as-is (LC_ALL=C); also guards the BSD awk abort
# when run locally on macOS.
commit "docs: éclair README"
# %, \ and & must survive verbatim (guards printf "$line" / awk -v regressions).
commit 'docs: handle 100%, a\b and a&b'
commit "chore: ignore local files"
commit "docs: update README"
commit "ci: bump checkout action"
commit "refactor: tidy helpers"
commit "fix: handle nil receiver"
commit "feat: add foo analyzer"

# Case 1: sections are classified, bullets capitalized, type prefixes dropped.
cat >expected <<'EOF'
## What's Changed

### Features
- Add foo analyzer

### Bug Fixes
- Handle nil receiver

### Other Changes
- Tidy helpers
- Bump checkout action
- Update README
- Ignore local files
- Handle 100%, a\b and a&b
- éclair README
- Initial import
EOF

if sh "$SCRIPT" HEAD >actual 2>err && diff -u expected actual; then
	pass "grouped, capitalized, prefix-stripped bullets"
else
	fail "grouped, capitalized, prefix-stripped bullets"
	cat err
fi

# Case 2: an empty range falls back to the maintenance line.
cat >expected <<'EOF'
## What's Changed

- Maintenance release
EOF

if sh "$SCRIPT" HEAD..HEAD >actual 2>err && diff -u expected actual; then
	pass "empty range prints maintenance release"
else
	fail "empty range prints maintenance release"
	cat err
fi

# Case 3: empty sections are omitted (newest commit is the lone feat:).
cat >expected <<'EOF'
## What's Changed

### Features
- Add foo analyzer
EOF

if sh "$SCRIPT" HEAD~1..HEAD >actual 2>err && diff -u expected actual; then
	pass "empty sections are omitted"
else
	fail "empty sections are omitted"
	cat err
fi

exit "$FAILED"
