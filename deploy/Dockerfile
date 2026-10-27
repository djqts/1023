ARG GO_IMAGE=golang:1.25-alpine
ARG RUNTIME_IMAGE=gcr.io/distroless/static-debian12:nonroot
ARG TARGETOS=linux
ARG TARGETARCH=amd64

FROM ${GO_IMAGE} AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/client ./cmd/client

FROM ${RUNTIME_IMAGE} AS server

ARG VERSION=dev
ARG COMMIT=unknown

LABEL org.opencontainers.image.title="banner-fingerprint-server" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}"

WORKDIR /app
COPY --from=build /out/server /app/server

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]

FROM ${RUNTIME_IMAGE} AS client

ARG VERSION=dev
ARG COMMIT=unknown

LABEL org.opencontainers.image.title="banner-fingerprint-client" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}"

WORKDIR /app
COPY --from=build /out/client /app/client

USER nonroot:nonroot
ENTRYPOINT ["/app/client"]
