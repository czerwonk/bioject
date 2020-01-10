FROM golang as builder
ADD . /go/bioject/
WORKDIR /go/bioject/cmd/bioject
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/bioject

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
