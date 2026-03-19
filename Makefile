proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/shipment.proto

test:
	go test ./... -v

run:
	go run ./cmd

build:
	go build -o bin/server ./cmd
