package grpcClient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ewohltman/grpc-example/pkg/filetransfer"
)

type Client struct {
	filetransfer.FileTransferClient
}

func (client *Client) UploadReader(ctx context.Context, reader io.Reader) {
	stream, err := client.Upload(ctx)
	if err != nil {
		log.Fatal(err)
	}

	input := os.Stdin

	var bytesRead int

	content := make([]byte, 1024)
	file := &filetransfer.File{}

	for {
		select {
		case <-stream.Context().Done():
			log.Fatal(stream.Context().Err())
		default:
			bytesRead, err = input.Read(content)
			if err != nil {
				if errors.Is(err, io.EOF) {
					response, err := stream.CloseAndRecv()
					if err != nil {
						log.Fatal(err)
					}

					fmt.Printf("Bytes written: %d\n", response.BytesWritten)

					return
				}

				log.Fatal(err)
			}

			file.Content = content[:bytesRead]

			err = stream.Send(file)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
