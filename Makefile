.PHONY: build build-install install eval lint lint-fast install-hook clean

## build: build the custom-gcl binary (requires golangci-lint v2.13.0)
build:
	golangci-lint custom

## build-install: build the binary directly via go build (without golangci-lint custom)
build-install:
	go build -o ./bin/custom-gcl ./cmd/custom-gcl

## install: install custom-gcl into $GOBIN via go install (for all projects)
install:
	go install ./cmd/custom-gcl

## eval: run evals for all rules (analysistest + testdata)
eval:
	go test ./...

## lint: rebuild custom-gcl and run it on the repository
lint: build lint-fast

## lint-fast: run the already built custom-gcl (no rebuild)
# --config is explicit on purpose: a plain run uses the config built into the
# binary, and this repository needs its own (testdata exclusions, its own
# giddirtree tree).
lint-fast:
	./bin/custom-gcl run --config .golangci.yml ./...

## install-hook: install a git pre-commit hook with the local check
install-hook:
	@printf '#!/bin/sh\n# Local rule gate (see RULES.md).\nexec make -s lint-fast\n' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "pre-commit hook installed (.git/hooks/pre-commit)"

clean:
	rm -rf bin
