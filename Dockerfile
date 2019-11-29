FROM golang as builder
RUN export GO111MODULE=on && go get github.com/czerwonk/bioject/cmd/bioject

FROM debian
ENV ZIPKIN_ENDPOINT ""
ENV DATA_PATH "/data"
ENV CONFIG_PATH "/config"
RUN mkdir /app && \
    mkdir /data
WORKDIR /app
COPY --from=builder /go/bin/bioject .
CMD ./bioject -config-file="$CONFIG_PATH/config.yml" -db-file="$DATA_PATH/routes.db" -zipkin-endpoint=$ZipkinEndpoint
VOLUME /config
EXPOSE 179
EXPOSE 1337
EXPOSE 9500
