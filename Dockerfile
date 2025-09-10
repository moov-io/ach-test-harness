# generated-from:8af0052ecb592976f405e5fa76b4f2ae824cb24e77859c243958226419efa06e DO NOT REMOVE, DO UPDATE

FROM golang:1.25 as builder
WORKDIR /src
ARG VERSION

RUN apt-get update && apt-get install -y make gcc g++ ca-certificates

COPY . .

RUN VERSION=${VERSION} make build

FROM debian:stable-slim AS runtime
LABEL maintainer="Moov <oss@moov.io>"

WORKDIR /

RUN apt-get update && apt-get install -y ca-certificates \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/bin/ach-test-harness /app/

ENV HTTP_PORT=2222
ENV HEALTH_PORT=3333

EXPOSE ${HTTP_PORT}/tcp
EXPOSE ${HEALTH_PORT}/tcp

VOLUME [ "/data", "/configs" ]

ENTRYPOINT ["/app/ach-test-harness"]
