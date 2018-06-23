.PHONY: all test clean man fast release install version info info_build info_vcs

GO15VENDOREXPERIMENT=1

PROG_NAME := "notify"

# dirs
DIST_DIR ?= ./dist
WRK_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# vcs
GIT_BRANCH := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))

# pkgs
SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')
PACKAGES = $(shell go list ./... | grep -v /vendor/)

# VERSION = $(shell cat "$(WRK_DIR)/VERSION" | tr '\n' '')
VERSION 			?= $(shell git describe --tags)
VERSION_INCODE 		= $(shell perl -ne '/^var version.*"([^"]+)".*$$/ && print "v$$1\n"' main.go)
VERSION_INCHANGELOG = $(shell perl -ne '/^\# Release (\d+(\.\d+)+) / && print "$$1\n"' CHANGELOG.md | head -n1)
VERSION_INFILE 		:= $(shell cat $(CURDIR)/VERSION)

VCS_GIT_REMOTE_URL 	= $(shell git config --get remote.origin.url)
VCS_GIT_VERSION 	?= $(VERSION)
VCS_GIT_DOMAIN 		:= "github.com"
VCS_GIT_OWNER 		:= "sniperkit"
VCS_GIT_PROJECT 	:= "snk.golang.notify"
VCS_GIT_REMOTE_URI	:= ${VCS_GIT_DOMAIN}/${VCS_GIT_OWNER}/${VCS_GIT_PROJECT}

CURBIN := $(shell which $(PROG_NAME))

all: deps test build install version

info: info_vs info_build

info_build:
	@echo "VERSION=${VERSION}"
	@echo "VERSION_INCODE=${VERSION_INCODE}"
	@echo "VERSION_INCHANGELOG=${VERSION_INCHANGELOG}"
	@echo "VERSION_INFILE=${VERSION_INFILE}"

info_vcs:
	@echo "VCS_GIT_REMOTE_URL=${VCS_GIT_REMOTE_URL}"
	@echo "VCS_GIT_VERSION=${VCS_GIT_VERSION}"

build:
	@go build -ldflags "-X $(VCS_GIT_REMOTE_URI)/pkg/version.VERSION=`$(VERSION_INFILE)`" -o ./bin/$(PROG_NAME) ./cmd/$(PROG_NAME)/*.go
	@./bin/$(PROG_NAME) --version

install: deps
	@go install-ldflags "-X $(VCS_GIT_REMOTE_URI)/pkg/version.VERSION=`$(VERSION_INFILE)`" ./cmd/$(PROG_NAME)
	@$(PROG_NAME) --version

fast: deps
	@go build -i -ldflags "-X main.VERSION=`cat VERSION`-dev" -o ./bin/$(PROG_NAME) ./cmd/$(PROG_NAME)/*.go
	@$(PROG_NAME) --version

deps:
	@go get -v -u github.com/mattn/goveralls
	@go get -v -u golang.org/x/tools/cmd/cover
	@glide install --strip-vendor

ci: tests coverage

tests:
	@go test ./pkg/... ./plugins/...

coverage:
	@go test -c -covermode=count -coverpkg=github.com/sniperkit/gotests/pkg,github.com/sniperkit/gotests/pkg/input,github.com/sniperkit/gotests/pkg/render,github.com/sniperkit/gotests/pkg/goparser,github.com/sniperkit/gotests/pkg/output,github.com/sniperkit/gotests/pkg/models
	@./gotests.test -test.coverprofile coverage.cov
	@goveralls -service=travis-ci -coverprofile=coverage.cov

clean:
	@go clean
	@rm -fr ./bin
	@rm -fr ./dist

release: build install
	@git tag -a `cat VERSION`
	@git push origin `cat VERSION`

