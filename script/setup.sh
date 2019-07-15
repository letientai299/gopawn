#!/usr/bin/env bash

# Source the common.sh script
# shellcheck source=./common.sh
. "$(git rev-parse --show-toplevel || echo ".")/script/common.sh"

cd "$PROJECT_DIR" || exit 1

# Mandatory tools
#-------------------------------------------------------------------------------

echo_info "Download golang dependencies"
go get -u

if ! has golangci-lint; then
  echo_info "Install golangci-lint for static code analysis (via curl)"
  curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh |
    sh -s -- -b "${GOPATH}/bin" v1.15.0 # last version to support go1.10
fi

if ! has goimports; then
  echo_info "Install goimports"
  go get -v -u golang.org/x/tools/cmd/goimports
fi

if ! has mockgen; then
  echo_info "Install mockgen"
  go get -v -u github.com/golang/mock/gomock
  go install -v -i github.com/golang/mock/mockgen
fi

if ! has mono; then
  if has brew; then
    echo_info "Install mono using homebrew"
    brew install mono
  else
    echo_warning "Mono is not found, please install it from https://www.mono-project.com/"
  fi
fi

# Nice to have tools, should only be installed when not on CI, to save build time
#-------------------------------------------------------------------------------
if [[ -z $CI ]]; then
  echo_info "Not on CI. Skipping nice to have tools"
else
  if ! has richgo; then
    echo_info "Install richgo for nicer go test output"
    go get -v -u github.com/kyoh86/richgo
  fi
fi

# Make the code ready for development
#-------------------------------------------------------------------------------

echo_info "Config git hooks push"
git config core.hooksPath "${PROJECT_DIR}/script/git-hooks"

cd "$WORKING_DIR" || exit 1
