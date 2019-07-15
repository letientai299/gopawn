# Self documented Makefile
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show list of make targets and their description
	@grep -E '^[%.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL:= help

.PHONY: setup
setup: ## Run setup script to prepare development environment
	@script/setup.sh

.PHONY: clean
clean: ## Clean project dir, remove build artifacts and logs
	@script/clean.sh

.PHONY: test
test: ## Generate mock and run all test. To run specified tests, use `./script/test.sh <pattern>`)
	@script/test.sh $*

.PHONY: lint
lint: ## Run linter
	@script/lint.sh

.PHONY: gen
gen: ## Show gen.sh help
	@script/gen.sh

gen.%: ## Gen target defined by %
	@script/gen.sh $*

.PHONY: build
build: ## Show build.sh help
	@script/build.sh

build.%: ## Build artifact defined by '%', e.g: 'make build.server` will trigger ./script/build.sh server
	@script/build.sh $*

all: clean setup gen.all build.all  ## Clean, setup, generate and then build all the binaries.
