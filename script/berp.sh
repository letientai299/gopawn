#!/usr/bin/env sh

. "$(git rev-parse --show-toplevel || echo ".")/script/common.sh"

cd "$PROJECT_DIR" || exit 1

mono ./tool/berp.exe $*

cd "$WORKING_DIR" || exit 1
