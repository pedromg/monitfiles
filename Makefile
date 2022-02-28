
.PHONY: all usage vet lint test build clean monitfiles go.toolchain.rev go.toolchain.ver VERSION.txt

BLD := $(go build)
GIT_HASH := $(shell git rev-parse HEAD)
BASE_HASH := $(shell git rev-list --max-count=1 HEAD -- VERSION.txt)
CHANGE_COUNT := $(shell git rev-list --count HEAD)
VER := '0.$(CHANGE_COUNT)'

usage:
	@echo "*** See Makefile for more details"

vet:
	@echo "*** vet'ing go code..."
	go vet ./...
	@echo "*** ... done vet!\n"

lint:
	@echo "*** lint'ing go code..."
	golint ./...
	@echo "*** ... done lint!\n"

test:
	@echo "*** Testing..."
	go test -v ./...
	@echo "*** ... done tests!\n"

build: test monitfiles
	@echo "*** ... done build!\n"

all: test vet lint monitfiles go.toolchain.ver go.toolchain.rev VERSION.txt
	@echo "*** ... done all builds! \n"

clean:
	@echo "*** Removing executables..."
	rm monitfiles
	@echo "*** ... done!\n"

monitfiles:
	go build

go.toolchain.rev:
	echo $(BASE_HASH) > go.toolchain.rev

go.toolchain.ver:
	go version > go.toolchain.ver

VERSION.txt:
	echo $(VER) > VERSION.txt
