FROM golang as builder
RUN go get github.com/czerwonk/bioject/cmd/bioject


FROM alpine:latest

RUN mkdir /app && \
    mkdir /data
WORKDIR /app

COPY --from=builder /go/bin/bioject .

CMD bioject -config-file=/config/config.yml

VOLUME /config

EXPOSE 179
EXPOSE 1337
