#!/bin/sh
set -eu
BASE_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO=${1:?usage: ROLLBACK.sh REPO_ROOT}
git -C "$REPO" apply --reverse --check "$BASE_DIR/DIFF_FILE"
git -C "$REPO" apply --reverse "$BASE_DIR/DIFF_FILE"
printf '%s\n' 'rollback applied'
