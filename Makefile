######################################################################
# @author      : annika (annika@berlin.ccc.de)
# @file        : Makefile
# @created     : Sunday Aug 16, 2020 19:24:54 CEST
######################################################################

test:
	go test ./pkg/...

update_openapi_static:
	go run ./cmd/b3scalectl export-openapi-schema > ./pkg/http/static/assets/docs/b3scale-openapi-v1.json


dev_static:
	cd ./cmd/b3scaled && CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) \
		go build -a -ldflags="-extldflags '-static'" .
