OUT_DIR=./out

.PHONY: server
server:
	mkdir -p ${OUT_DIR}
	go build -o ${OUT_DIR}/bioject ./cmd/bioject/main.go

.PHONY: client
client:
	mkdir -p ${OUT_DIR}
	go build -o ${OUT_DIR}/biojecter ./cmd/biojecter/main.go

.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false --go-grpc_out=. ./proto/*.proto
