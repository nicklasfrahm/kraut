FROM golang:1.21 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests.
COPY go.mod go.mod
COPY go.sum go.sum
# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer.
RUN go mod download

# Copy the go source.
COPY Makefile Makefile
COPY cmd/ cmd/
COPY api/ api/
COPY pkg/ pkg/
COPY internal/ internal/
COPY .git/ .git/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
ARG VERSION
ARG COMPONENT
RUN VERSION=${VERSION} make bin/${COMPONENT} && mv bin/${COMPONENT} app

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
ARG VERSION
ARG COMPONENT

# Reference: https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.authors "Nicklas Frahm <nicklas.frahm@gmail.com>"
LABEL org.opencontainers.image.documentation "https://kraut.nicklasfrahm.dev"
LABEL org.opencontainers.image.source https://github.com/nicklasfrahm/kraut
LABEL org.opencontainers.image.vendor "Nicklas Frahm"
LABEL org.opencontainers.image.licenses "MIT"
LABEL org.opencontainers.image.description "kraut is an infrastructure orchestrator built on top of Kubernetes."
LABEL org.opencontainers.image.url "https://github.com/nicklasfrahm/kraut/pkgs/container/kraut%2F${COMPONENT}"
LABEL org.opencontainers.image.title "kraut-${COMPONENT}"
LABEL org.opencontainers.image.version ${VERSION}

WORKDIR /
COPY --from=builder /workspace/app .
USER 65532:65532

ENTRYPOINT ["/app"]
