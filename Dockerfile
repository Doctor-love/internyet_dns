# Builder container
FROM golang:latest AS builder

WORKDIR /go/src
COPY main.go .

## Ensures that built binary is static
RUN GO111MODULE=auto CGO_ENABLED=1 GOOS=linux go build \
	-a -tags netgo -ldflags '-w -extldflags "-static"' \
	-o internyet_dns

# Runtime container
FROM scratch

WORKDIR /
COPY --from=builder /go/src/internyet_dns .

USER 1000
ENTRYPOINT ["/internyet_dns"]
