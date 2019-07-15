#!/usr/bin/env bash

# Source the common.sh script
# shellcheck source=./common.sh
. "$(git rev-parse --show-toplevel || echo ".")/script/common.sh"

usage() {
  cat <<EOF
Generate code and other artifacts that required to build binaries. Known recipes:

  parser      generate parser from defined grammar and template, (need special setup)

See "make build" if you are looking building binary artifacts.
EOF
}

gen_parser() {
  echo_info "Generate parser code grammar and razor template"
  ./script/berp.sh --grammar ./grammar/gherkin.berp \
    --template ./grammar/parser.go.razor \
    --output ./internal/gherkin/parser.go
}

cd "$PROJECT_DIR" || exit 1

case "$1" in
parser)
  gen_parser
  exit
  ;;
*)
  usage
  exit
  ;;
esac

cd "$WORKING_DIR" || exit 1
