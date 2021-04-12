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

LDFLAGS := -X gitlab.com/infra.run/public/b3scale/pkg/config.Version=$(VERSION) \
		   -X gitlab.com/infra.run/public/b3scale/pkg/config.Build=$(BUILD)
LDFLAGS_STATIC := $(LDFLAGS) -extldflags "-static"


all: b3scaled b3scalectl b3scalenoded

static: b3scaled_static b3scalectl_static b3scalenoded_static

b3scaled:
	cd cmd/b3scaled && go build $(CFLAGS) -ldflags '$(LDFLAGS)'

b3scalectl:
	cd cmd/b3scalectl && go build $(CFLAGS) -ldflags '$(LDFLAGS)'

b3scalenoded:
	cd cmd/b3scalenoded && go build $(CFLAGS) -ldflags '$(LDFLAGS)'

b3scaled_static:
	cd cmd/b3scaled && CGO_ENABLED=0 GOOS=linux go build $(CFLAGS) -a -ldflags '$(LDFLAGS_STATIC)'

b3scalectl_static:
	cd cmd/b3scalectl && CGO_ENABLED=0 GOOS=linux go build $(CFLAGS) -a -ldflags '$(LDFLAGS_STATIC)'

b3scalenoded_static:
	cd cmd/b3scalenoded && CGO_ENABLED=0 GOOS=linux go build $(CFLAGS) -a -ldflags '$(LDFLAGS_STATIC)'


.PHONY: clean test

test:
	cd pkg/cluster && go test
	cd pkg/config && go test
	cd pkg/store && go test
	cd pkg/bbb && go test
	cd pkg/iface/http && go test
	cd pkg/middlewares/requests && go test
	cd pkg/middlewares/routing && go test
	cd cmd/b3scalectl && go test

clean:
	rm -f cmd/b3scaled/b3scaled
	rm -f cmd/b3scalectl/b3scalectl
	rm -f cmd/b3scalectl/b3scalenoded

