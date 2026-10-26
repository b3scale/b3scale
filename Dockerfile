FROM alpine:3.24.3 AS builder

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY /b3scaled /

ENTRYPOINT ["/b3scaled"]
