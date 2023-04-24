FROM golang as builder
ADD . /go/bioject/
WORKDIR /go/bioject/cmd/bioject
RUN GOOS=linux go build -o /go/bin/bioject

FROM debian
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
