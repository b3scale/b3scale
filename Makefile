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


all: b3scaled b3scalectl b3scalenoded b3scaleagent

static: b3scaled_static b3scalectl_static b3scalenoded_static b3scaleagent_static

b3scaled:
	cd cmd/b3scaled && go build $(CFLAGS) -ldflags '$(LDFLAGS)'

b3scalectl:
	cd cmd/b3scalectl && go build $(CFLAGS) -ldflags '$(LDFLAGS)'

b3scalenoded:
	cd cmd/b3scalenoded && go build $(CFLAGS) -ldflags '$(LDFLAGS)'

b3scaleagent:
	cd cmd/b3scaleagent && go build $(CFLAGS) -ldflags '$(LDFLAGS)'

b3scaled_static:
	cd cmd/b3scaled && CGO_ENABLED=0 GOOS=linux go build $(CFLAGS) -a -ldflags '$(LDFLAGS_STATIC)'

b3scalectl_static:
	cd cmd/b3scalectl && CGO_ENABLED=0 GOOS=linux go build $(CFLAGS) -a -ldflags '$(LDFLAGS_STATIC)'

b3scalenoded_static:
	cd cmd/b3scalenoded && CGO_ENABLED=0 GOOS=linux go build $(CFLAGS) -a -ldflags '$(LDFLAGS_STATIC)'

b3scaleagent_static:
	cd cmd/b3scaleagent && CGO_ENABLED=0 GOOS=linux go build $(CFLAGS) -a -ldflags '$(LDFLAGS_STATIC)'


.PHONY: clean test

test:
	go test ./pkg/...

clean:
	rm -f cmd/b3scaled/b3scaled
	rm -f cmd/b3scalectl/b3scalectl
	rm -f cmd/b3scalectl/b3scalenoded

