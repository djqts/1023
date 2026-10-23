# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/client ./cmd/client

FROM gcr.io/distroless/static-debian12:nonroot AS server

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

FROM gcr.io/distroless/static-debian12:nonroot AS client

ARG VERSION=dev
ARG COMMIT=unknown

LABEL org.opencontainers.image.title="banner-fingerprint-client" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}"

WORKDIR /app
COPY --from=build /out/client /app/client

USER nonroot:nonroot
ENTRYPOINT ["/app/client"]
