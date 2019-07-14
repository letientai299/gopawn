#!/usr/bin/env bash

# Source the common.sh script
# shellcheck source=./common.sh
. "$(git rev-parse --show-toplevel || echo ".")/scripts/common.sh"

cd "$PROJECT_DIR" || exit 1

echo_info "Remove all log files"
find . -name '*.log' -type f -print | xargs rm -vf
echo_info "Remove all empty log folders"
# This command remove empty folder that is children of a "log" folder
find . -path '*/logfile/*' -type d -empty -print | xargs rm -vrf
find . -path '*/log/*' -type d -empty -print | xargs rm -vrf
# And now we remove all the remaining empty "log" folders
find . -path '*/log' -type d -empty -print | xargs rm -vrf

echo_info "Remove binary artifacts"
rm -vrf ./bin

cd "$WORKING_DIR" || exit 1
