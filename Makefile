.PHONY: generate lint test build

generate:
	protoc -I pkg/filetransfer/ pkg/filetransfer/filetransfer.proto --go_out=plugins=grpc:pkg/filetransfer
	goimports -w pkg/filetransfer/filetransfer.pb.go

lint:
	golangci-lint run ./...

build:
	CGO_ENABLED=0 go build -o build/package/grpc-client/grpc-client cmd/grpc-client/grpc-client.go
	CGO_ENABLED=0 go build -o build/package/grpc-server/grpc-server cmd/grpc-server/grpc-server.go
