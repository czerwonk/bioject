FROM golang:1.26.6-alpine3.24@sha256:af8d6740070b8906d12eae1c3e3ea0957fb63f492051ea05e354c38ef9fe88df AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

ADD . /go/bioject/
WORKDIR /go/bioject/cmd/bioject
RUN CGO_ENABLED=1 GOOS=linux go build \
	-ldflags="-linkmode external -extldflags -static" \
	-o /go/bin/bioject

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
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
