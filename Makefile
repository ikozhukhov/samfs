NAME := samfs
DESC := a nice toolkit of helpful things
PREFIX ?= $(PWD)
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
PKG_RELEASE ?= 1
PROJECT_URL := "https://github.com/smihir/$(NAME)"
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.buildTime=$(BUILDTIME)' \
           -X 'main.builder=$(BUILDER)' \
           -X 'main.goversion=$(GOVERSION)'

# development tasks
test:
	go test $$(go list ./... | grep -v /vendor/)


PACKAGES := $(shell find ./* -type d | grep -v vendor)

coverage:
	@echo "mode: set" > cover.out
	@for package in $(PACKAGES); do \
        if ls $${package}/*.go &> /dev/null; then  \
        go test -coverprofile=$${package}/profile.out $${package}; fi; \
        if test -f $${package}/profile.out; then \
        cat $${package}/profile.out | grep -v "mode: set" >> cover.out; fi; \
    done
	@-go tool cover -html=cover.out -o cover.html

benchmark:
	@echo "Running tests..."
	@go test -bench=. $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)

CMD_SOURCES := $(shell find cmd -name main.go)
TARGETS := $(patsubst cmd/%/main.go,%,$(CMD_SOURCES))

%: cmd/%/main.go
	go build -ldflags "$(LDFLAGS)" -o $@ $<

lint:
	go vet $$(go list ./... | grep -v /vendor/)

INSTALLED_TARGETS = $(addprefix $(PREFIX)/bin/, $(TARGETS))

$(PREFIX)/bin/%: %
	mv $< $@

install_location: $(PREIX)/bin
	mkdir -p $(PREFIX)/bin

install: install_location $(INSTALLED_TARGETS)

clean:
	rm -rf bin

all: lint $(TARGETS) install
.DEFAULT_GOAL:=all