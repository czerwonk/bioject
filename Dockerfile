FROM golang:1.27.4-alpine3.24@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

ADD . /go/bioject/
WORKDIR /go/bioject/cmd/bioject
RUN CGO_ENABLED=1 GOOS=linux go build \
	-ldflags="-linkmode external -extldflags -static" \
	-o /go/bin/bioject

FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf
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
