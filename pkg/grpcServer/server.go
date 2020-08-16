package grpcServer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/ewohltman/grpc-example/pkg/filetransfer"
)

type Server struct{}

func (server Server) Upload(stream filetransfer.FileTransfer_UploadServer) (err error) {
	log.Printf("Receiving upload stream\n")

	var (
		inputFile         *filetransfer.File
		outputFile        *os.File
		totalBytesWritten int64
	)

	outputFile, err = os.OpenFile(filepath.Clean("/tmp/temp-file"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		log.Printf("error opening temp file: %s\n", err)
		return fmt.Errorf("error opening temp file: %w", err)
	}

	defer func() {
		closeErr := outputFile.Close()
		if closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%s: %w", closeErr, err)
			} else {
				err = closeErr
			}

			log.Printf("%s\n", err)
		}
	}()

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("%s\n", err)
			return stream.Context().Err()
		default:
			inputFile, err = stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return stream.SendAndClose(&filetransfer.UploadResponse{BytesWritten: totalBytesWritten})
				}
				log.Printf("error receiving input stream: %s\n", err)
				return fmt.Errorf("error receiving input stream: %w", err)
			}

			var bytesWritten int64

			bytesWritten, err = io.Copy(outputFile, bytes.NewReader(inputFile.Content))
			if err != nil {
				log.Printf("error copying input stream contents: %s\n", err)
				return fmt.Errorf("error copying input stream contents: %w", err)
			}

			totalBytesWritten += bytesWritten
		}
	}
}
