######################################################################
# @author      : annika (annika@berlin.ccc.de)
# @file        : Makefile
# @created     : Sunday Aug 16, 2020 19:24:54 CEST
######################################################################

# Force using the vendored dependencies
VENDOR := false

# Set the release version
VERSION := $(shell git tag --points-at HEAD)
ifeq ($(VERSION),)
  VERSION=HEAD
endif

# Set the release build
BUILD := $(shell git rev-parse --short HEAD)


CFLAGS := -buildmode=pie
ifneq ($(VENDOR), false)
  CFLAGS += -mod=vendor
endif

LDFLAGS := -X github.com/b3scale/b3scale/pkg/config.Version=$(VERSION) \
		   -X github.com/b3scale/b3scale/pkg/config.Build=$(BUILD)
LDFLAGS_STATIC := $(LDFLAGS) -extldflags "-static"

