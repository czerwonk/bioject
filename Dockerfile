FROM golang as builder
RUN go get github.com/czerwonk/bioject/cmd/bioject

FROM debian
RUN mkdir /app && \
    mkdir /data
WORKDIR /app
COPY --from=builder /go/bin/bioject .
CMD ./bioject -config-file=/config/config.yml -db-file=/config/routes.db
VOLUME /config
EXPOSE 179
EXPOSE 1337
EXPOSE 9500
