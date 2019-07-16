#!/usr/bin/env bash

# Source the common.sh script
# shellcheck source=./common.sh
. "$(git rev-parse --show-toplevel || echo ".")/script/common.sh"

usage() {
  cat <<EOF
Generate code and other artifacts that required to build binaries. Known recipes:
  gogen       run go generate

  parser      generate parser from defined grammar and template, (need special setup)

  proto       generate golang code from the proto msg

See "make build" if you are looking building binary artifacts.
EOF
}

gen_all() {
  gen_gogen
  gen_proto
  gen_parser
}

gen_gogen(){
  go generate ./...
}

gen_proto() {
  echo_info "Generate golang code from proto"
  protoc --go_out=internal/msg/ \
    --proto_path=./internal/msg \
    msg.proto
}

gen_parser() {
  echo_info "Generate parser code grammar and razor template"
  ./script/berp.sh --grammar ./grammar/gherkin.berp \
    --template ./grammar/parser.go.razor \
    --output ./internal/gherkin/parser.go
}

cd "$PROJECT_DIR" || exit 1

case "$1" in
all)
  gen_all
  exit
  ;;
gogen)
  gen_gogen
  exit
  ;;
proto)
  gen_proto
  exit
  ;;
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
