FROM golang as builder
RUN go get github.com/czerwonk/bioject/cmd/bioject

FROM debian
ENV ZipkinEndpoint ""
RUN mkdir /app && \
    mkdir /data
WORKDIR /app
COPY --from=builder /go/bin/bioject .
CMD ./bioject -config-file=/config/config.yml -db-file=/config/routes.db -zipkin-endpoint=$ZipkinEndpoint
VOLUME /config
EXPOSE 179
EXPOSE 1337
EXPOSE 9500
