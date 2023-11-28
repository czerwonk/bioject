#!/bin/sh
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false --go-grpc_out=. *.proto
