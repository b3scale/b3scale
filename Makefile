######################################################################
# @author      : annika (annika@berlin.ccc.de)
# @file        : Makefile
# @created     : Sunday Aug 16, 2020 19:24:54 CEST
######################################################################

all: b3scaled b3scalectl

b3scaled:
	cd cmd/b3scaled && go build

b3scalectl:
	cd cmd/b3scalectl && go build

.PHONY: clean test

test:
	cd pkg/cluster && go test
	cd pkg/config && go test
	cd pkg/store && go test
	cd pkg/bbb && go test
	cd pkg/iface/http && go test
	cd pkg/middlewares/routing && go test
	cd pkg/middlewares/requests && go test 

clean:
	rm -f cmd/b3scaled/b3scaled
	rm -f cmd/b3scalectl/b3scalectl

