FROM golang:1.27.2-alpine3.24@sha256:85dc1069ac644ea3c527b177303a406eb3358192816cd7f9e5848eb658851673 AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

ADD . /go/bioject/
WORKDIR /go/bioject/cmd/bioject
RUN CGO_ENABLED=1 GOOS=linux go build \
	-ldflags="-linkmode external -extldflags -static" \
	-o /go/bin/bioject

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
ENV ZIPKIN_ENDPOINT ""
ENV DATA_PATH "/data"
ENV CONFIG_PATH "/config"
RUN mkdir /app && \
    mkdir /data
WORKDIR /app
COPY --from=builder /go/bin/bioject .
CMD ./bioject -config-file="$CONFIG_PATH/config.yml" -db-file="$DATA_PATH/routes.db"
VOLUME /config
EXPOSE 179
EXPOSE 1337
EXPOSE 9500
