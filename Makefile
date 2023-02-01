.DEFAULT_GOAL := help
CURRENTTAG:=$(shell git describe --tags --abbrev=0)
NEWTAG ?= $(shell bash -c 'read -p "Please provide a new tag (currnet tag - ${CURRENTTAG}): " newtag; echo $$newtag')
GOFLAGS=-mod=mod
GOPRIVATE=github.com/threeport/*,github.com/qleet/*

#help: @ List available tasks
help:
	@clear
	@echo "Usage: make COMMAND"
	@echo "Commands :"
	@grep -E '[a-zA-Z\.\-]+:.*?@ .*$$' $(MAKEFILE_LIST)| tr -d '#' | awk 'BEGIN {FS = ":.*?@ "}; {printf "\033[32m%-19s\033[0m - %s\n", $$1, $$2}'

#clean: @ Cleanup
clean:
	@rm -rf ./dist
	@rm -rf ./completions
	@rm -f ./tptctl

#get: @ Download and install dependency packages
get:
	@export GOPRIVATE=$(GOPRIVATE); export GOFLAGS=$(GOFLAGS); go get . ; go mod tidy

#update: @ Update dependencies to latest versions
update:
	@export GOPRIVATE=$(GOPRIVATE); export GOFLAGS=$(GOFLAGS); go get -u; go mod tidy

#test: @ Run tests
test:
	@export GOPRIVATE=$(GOPRIVATE); go generate
	@export GOPRIVATE=$(GOPRIVATE); export GOFLAGS=$(GOFLAGS); go test $(go list ./... | grep -v /internal/setup)

#build: @ Build tptctl binary
build:
	@export GOPRIVATE=$(GOPRIVATE); go generate
	@export GOPRIVATE=$(GOPRIVATE); export GOFLAGS=$(GOFLAGS); export CGO_ENABLED=0; go build -a -o tptctl main.go

#install: @ Install the tptctl CLI
install: build
	sudo mv ./tptctl /usr/local/bin/

#release: @ Create and push a new tag
release: build
	$(eval NT=$(NEWTAG))
	@echo -n "Are you sure to create and push ${NT} tag? [y/N] " && read ans && [ $${ans:-N} = y ]
	@echo ${NT} > ./cmd/version.txt
	@git add -A
	@git commit -a -s -m "Cut ${NT} release"
	@git tag -a -m "Cut ${NT} release" ${NT}
	@git push origin ${NT}
	@git push
	@echo "Done."

#test-release-local: @ Build binaries locally without publishing
test-release-local: clean
	@goreleaser check
	@goreleaser release --rm-dist --snapshot

#version: @ Print current version(tag)
version:
	@echo $(shell git describe --tags --abbrev=0)

#codegen-subcommand:  @ Build subcommand - a tool for generating subcommand source code
codegen-subcommand:
	@go build -o bin/subcommand codegen/subcommand/main.go
