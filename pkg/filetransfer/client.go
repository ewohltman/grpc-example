package filetransfer

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Client struct {
	FileTransferClient
	log        *logrus.Entry
	connection *grpc.ClientConn
}

func NewClient(log *logrus.Entry, serverAddress string) (*Client, error) {
	connection, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("error dialing gRPC server: %w", err)
	}

	client := &Client{
		FileTransferClient: NewFileTransferClient(connection),
		log:                log,
		connection:         connection,
	}

	return client, nil
}

func (client *Client) UploadReader(ctx context.Context, fileName string, input io.Reader) {
	stream, err := client.Upload(ctx)
	if err != nil {
		client.log.WithError(err).Error("Error creating upload stream")
		return
	}

	var bytesRead int

	file := &File{Name: fileName}
	content := make([]byte, 1024)

	for {
		select {
		case <-stream.Context().Done():
			client.log.WithError(stream.Context().Err()).Error("Upload stream context closed")
		default:
			bytesRead, err = input.Read(content)
			if err != nil {
				if errors.Is(err, io.EOF) {
					response, err := stream.CloseAndRecv()
					if err != nil {
						client.log.WithError(err).Error("Error closing upload stream and getting response")
						return
					}

					client.log.WithField("bytes", response.BytesWritten).Info("Upload complete")
					return
				}

				client.log.WithError(err).Error("Error reading input file")
				return
			}

			file.Content = content[:bytesRead]

			err = stream.Send(file)
			if err != nil {
				client.log.WithError(err).Error("Error sending input file")
				return
			}
		}
	}
}

func (client *Client) Close() error {
	return client.connection.Close()
}
