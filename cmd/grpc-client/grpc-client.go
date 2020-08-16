package main

import (
	"context"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"

	"github.com/ewohltman/grpc-example/pkg/filetransfer"
	"github.com/ewohltman/grpc-example/pkg/grpcClient"
)

const serverAddress = "localhost:8080"

func handleClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Printf("Error: %s\n", err)
	}
}

func main() {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}

	defer handleClose(conn)

	client := &grpcClient.Client{
		FileTransferClient: filetransfer.NewFileTransferClient(conn),
	}

	client.UploadReader(context.Background(), os.Stdin)
}
