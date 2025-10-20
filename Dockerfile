# Build the manager binary
FROM golang:1.24 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY conf/ conf/
COPY exporter/ exporter/
COPY storage/ storage/
COPY vault/ vault/
COPY encryption/ encryption/

# Build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o vault-manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:nonroot
FROM alpine:edge
WORKDIR /
COPY --from=builder /workspace/vault-manager .

RUN addgroup -g 1001 vaultmanager
RUN adduser -u 1001 -G vaultmanager -s /sbin/nologin -D vaultmanager

USER vaultmanager
RUN mkdir -pv /home/vaultmanager/data

ENTRYPOINT ["/vault-manager"]
