FROM golang as builder
RUN go get -d -v github.com/czerwonk/bioject/cmd/bioject
WORKDIR /go/src/github.com/czerwonk/bioject/cmd/bioject
RUN GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir /app && \
    mkdir /data
WORKDIR /app
COPY --from=builder /go/src/github.com/czerwonk/bioject/cmd/bioject/app bioject
CMD ./bioject -config-file=/config/config.yml
VOLUME /config
EXPOSE 179
EXPOSE 1337
