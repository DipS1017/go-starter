ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,source=go.mod,target=go.mod \
  --mount=type=bind,source=go.sum,target=go.sum \
  go mod download -x

COPY . .

RUN mkdir -p /out/build

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod/ CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/build/server main.go

FROM alpine:3.21 AS prod

WORKDIR /app

LABEL org.opencontainers.image.title="DBA api"

RUN apk --update add \
  ca-certificates \
  tzdata \
  --no-cache \
  && \
  update-ca-certificates

COPY --from=builder /out/build/server /app/server

ENV PORT=8080

EXPOSE ${PORT}

CMD ["/app/server"]
