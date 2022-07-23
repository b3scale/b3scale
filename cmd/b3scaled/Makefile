######################################################################
# @author      : annika (annika@berlin.ccc.de)
# @file        : Makefile
# @created     : Sunday Aug 16, 2020 19:24:54 CEST
######################################################################

include ../../Makefile.inc

BIN=$(shell basename $(CURDIR))

all: build

build: *.go
	go build $(CFLAGS) -ldflags '$(LDFLAGS)' -o ../../build/$(BIN)

static: *.go
	CGO_ENABLED=0 go build $(CFLAGS) -a -ldflags '$(LDFLAGS_STATIC)' -o ../../build/$(BIN)-static

.PHONY: clean

clean:
	rm -f ../../build/$(BIN)
	rm -f ../../build/$(BIN)-static
