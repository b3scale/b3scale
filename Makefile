######################################################################
# @author      : annika (annika@berlin.ccc.de)
# @file        : Makefile
# @created     : Sunday Aug 16, 2020 19:24:54 CEST
######################################################################

all: b3scaled

b3scaled:
	cd cmd/b3scaled && go build

.PHONY: clean test

test:
	cd pkg/cluster && go test -v
	cd pkg/config && go test -v
	cd pkg/bbb && go test -v
	cd pkg/iface/http && go test -v
	# cd pkg/middlewares/routing && go test -v
	cd pkg/middlewares/requests && go test -v

clean:
	rm -f cmd/b3scaled/b3scaled
	rm -f cmd/b3scalectl/b3scalectl

