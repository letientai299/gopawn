#!/usr/bin/env bash

# Source the common.sh script
# shellcheck source=./common.sh
. "$(git rev-parse --show-toplevel || echo ".")/script/common.sh"

build() {
	cmd=$1
	echo_info "Building $cmd"
	target=$(echo $cmd | sed 's/cmd/bin/')
	go build -i -v -o $target $cmd
	ls -lah -d $target
}

build_all() {
	# Build
	for cmd_package in $(find ./cmd -type d); do
		# skip the folder if there's no go file in it
		ls "$cmd_package"/*.go >/dev/null 2>&1 || continue
		# build the cmd
		build "$cmd_package"
	done
}

check_target() {
	cmd="$1"
	if [ -z $1 ]; then
		usage
		exit
	fi

	target="./cmd/$cmd"
	if [ -d "$target" ]; then
		build "$target"
	else
		echo_warn "$cmd is not exists under 'cmd'. Here's how to build script"
		echo
		usage
	fi
}

usage() {
	cat <<EOF
Build binary artifacts this service. Build target could be any folder under the
'cmd' one, such as:

./script/build.sh gopawn

Or, if you trigger this script via make:

make build.gopawn

You can also build all packages under 'cmd' into binaries by a special target:
'all'.

make build.all

To do cross compilation, such as build binary for Linux while working on Mac

  GOOS=linux make build.server
EOF
}

cd "$PROJECT_DIR" || exit 1

case "$1" in
all)
	build_all
	exit
	;;
-h)
	usage
	exit
	;;
*)
	check_target "$1"
	exit
	;;
esac

cd "$WORKING_DIR" || exit 1
