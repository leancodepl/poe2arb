#!/usr/bin/env bash
#
# Integration test for `poe2arb seed` and `poe2arb upload` against a real
# POEditor project. The project is wiped clean at the start and end of the run
# so the test is repeatable.
#
# Required environment variables:
#   POEDITOR_TOKEN          POEditor API token with WRITE access
#   POE_PROJECT_ID          ID of a dedicated test POEditor project (do NOT use
#                           a real project — its terms WILL be deleted)
#
# Run from the integration_test/ directory.

set -euo pipefail

: "${POEDITOR_TOKEN:?POEDITOR_TOKEN must be set}"
: "${POE_PROJECT_ID:?POE_PROJECT_ID must be set}"

API="https://api.poeditor.com/v2"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="${SCRIPT_DIR}/../write_test_data"
ARB_DIR="lib/l10n_write_test"
TMP_DIR="$(mktemp -d)"

cleanup() {
  echo "Wiping test project terms..."
  curl -fsS -X POST "${API}/projects/sync" \
    --data-urlencode "api_token=${POEDITOR_TOKEN}" \
    --data-urlencode "id=${POE_PROJECT_ID}" \
    --data-urlencode "data=[]" \
    | jq -e '.response.code == "200"' >/dev/null
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

step() {
  echo
  echo "============================================================"
  echo "STEP: $*"
  echo "============================================================"
}

# Ensure our test project really is empty before we start.
echo "Pre-test cleanup..."
curl -fsS -X POST "${API}/projects/sync" \
  --data-urlencode "api_token=${POEDITOR_TOKEN}" \
  --data-urlencode "id=${POE_PROJECT_ID}" \
  --data-urlencode "data=[]" \
  | jq -e '.response.code == "200"' >/dev/null

reset_arb_dir() {
  rm -rf "${ARB_DIR}"
  mkdir -p "${ARB_DIR}"
}

stage_files() {
  reset_arb_dir
  cp "${DATA_DIR}/$1/"*.arb "${ARB_DIR}/"
}

run_dry_run() {
  local out="$1"
  poe2arb upload --dry-run -p "${POE_PROJECT_ID}" -o "${ARB_DIR}" 2>&1 | tee "${out}"
}

assert_grep() {
  local pattern="$1"
  local file="$2"
  if ! grep -Eq "$pattern" "$file"; then
    echo "FAIL: expected /$pattern/ in $file but did not find it" >&2
    echo "----- $file -----" >&2
    cat "$file" >&2
    echo "------------------" >&2
    exit 1
  fi
}

assert_no_grep() {
  local pattern="$1"
  local file="$2"
  if grep -Eq "$pattern" "$file"; then
    echo "FAIL: did not expect /$pattern/ in $file but found it" >&2
    echo "----- $file -----" >&2
    cat "$file" >&2
    echo "------------------" >&2
    exit 1
  fi
}

# --------------------------------------------------------------------
# Stage 1: seed an empty project
# --------------------------------------------------------------------
step "seed empty project from stage1 ARBs"
stage_files stage1
poe2arb seed -p "${POE_PROJECT_ID}" -o "${ARB_DIR}"

step "verify: dry-run after seed should report zero changes"
run_dry_run "${TMP_DIR}/dry_after_seed.txt"
assert_grep "term additions: 0"  "${TMP_DIR}/dry_after_seed.txt"
assert_grep "term deletions: 0"  "${TMP_DIR}/dry_after_seed.txt"
assert_no_grep "added,"          "${TMP_DIR}/dry_after_seed.txt"
assert_no_grep "updated,"        "${TMP_DIR}/dry_after_seed.txt"

# --------------------------------------------------------------------
# Stage 2: add a new term and update an existing one
# --------------------------------------------------------------------
step "modify ARBs (stage2: add 'welcome', update 'hello')"
stage_files stage2

step "dry-run should announce 1 addition and translation updates"
run_dry_run "${TMP_DIR}/dry_stage2_pre.txt"
assert_grep "term additions: 1"          "${TMP_DIR}/dry_stage2_pre.txt"
assert_grep "term deletions: 0"          "${TMP_DIR}/dry_stage2_pre.txt"
assert_grep "\\+ welcome"                "${TMP_DIR}/dry_stage2_pre.txt"
# both languages should report the update of 'hello'
assert_grep "en:.*updated"               "${TMP_DIR}/dry_stage2_pre.txt"
assert_grep "pl:.*updated"               "${TMP_DIR}/dry_stage2_pre.txt"

step "actual upload (no --force needed: no deletions)"
poe2arb upload -p "${POE_PROJECT_ID}" -o "${ARB_DIR}"

step "verify: dry-run after upload should report zero changes"
run_dry_run "${TMP_DIR}/dry_stage2_post.txt"
assert_grep "term additions: 0"  "${TMP_DIR}/dry_stage2_post.txt"
assert_grep "term deletions: 0"  "${TMP_DIR}/dry_stage2_post.txt"
assert_no_grep "added,"          "${TMP_DIR}/dry_stage2_post.txt"
assert_no_grep "updated,"        "${TMP_DIR}/dry_stage2_post.txt"

# --------------------------------------------------------------------
# Stage 3: remove a term locally → must abort without --force
# --------------------------------------------------------------------
step "remove a term locally (stage3: drop 'thanks')"
stage_files stage3

step "dry-run should announce the pending deletion"
run_dry_run "${TMP_DIR}/dry_stage3_pre.txt"
assert_grep "term deletions: 1" "${TMP_DIR}/dry_stage3_pre.txt"
assert_grep "thanks"            "${TMP_DIR}/dry_stage3_pre.txt"

step "upload without --force must fail"
if poe2arb upload -p "${POE_PROJECT_ID}" -o "${ARB_DIR}"; then
  echo "FAIL: upload without --force should have errored on deletions" >&2
  exit 1
fi

step "upload --force must succeed and delete the term"
poe2arb upload --force -p "${POE_PROJECT_ID}" -o "${ARB_DIR}"

step "verify: dry-run after force-upload should report zero changes"
run_dry_run "${TMP_DIR}/dry_stage3_post.txt"
assert_grep "term additions: 0"  "${TMP_DIR}/dry_stage3_post.txt"
assert_grep "term deletions: 0"  "${TMP_DIR}/dry_stage3_post.txt"

step "verify via fresh poe download: 'thanks' is gone"
reset_arb_dir
poe2arb poe -p "${POE_PROJECT_ID}" -o "${ARB_DIR}"
if grep -q "thanks" "${ARB_DIR}/app_en.arb"; then
  echo "FAIL: 'thanks' should have been deleted but is still present" >&2
  cat "${ARB_DIR}/app_en.arb" >&2
  exit 1
fi
assert_grep '"hello": "Hello there"' "${ARB_DIR}/app_en.arb"
assert_grep '"welcome": "Welcome"'   "${ARB_DIR}/app_en.arb"

echo
echo "============================================================"
echo "All write integration tests passed."
echo "============================================================"
