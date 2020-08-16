package filetransfer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"

	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
)

const (
	outputFileFlags       = os.O_CREATE | os.O_TRUNC | os.O_RDWR
	outputFilePermissions = 0644
)

type ServerConfig struct {
	Log             *logrus.Entry
	Network         string
	Address         string
	OutputDirectory string
}

type Server struct {
	log             *logrus.Entry
	outputDirectory string
}

func NewServer(config *ServerConfig) (gracefulStop func(), err error) {
	var listener net.Listener

	listener, err = net.Listen(config.Network, config.Address)
	if err != nil {
		return nil, fmt.Errorf("error listening for %s on %s: %w", config.Network, config.Address, err)
	}

	grpcServer := grpc.NewServer()
	server := &Server{log: config.Log, outputDirectory: config.OutputDirectory}

	RegisterFileTransferServer(grpcServer, server)

	go func() {
		err := grpcServer.Serve(listener)
		if err != nil {
			server.log.WithError(err).Errorf("gRPC server error")
		}
	}()

	return grpcServer.GracefulStop, nil
}

func (server *Server) Upload(stream FileTransfer_UploadServer) (err error) {
	server.log.Info("Receiving upload stream")

	var (
		inputFile         *File
		totalBytesWritten int64
		once              sync.Once
	)

	outputFile := &os.File{}

	defer func() {
		if outputFile != nil {
			closeErr := outputFile.Close()
			if closeErr != nil {
				if err != nil {
					err = fmt.Errorf("%s: %w", closeErr, err)
				} else {
					err = closeErr
				}
			}
		}
	}()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		default:
			inputFile, err = stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					server.log.Info("Upload complete")
					return stream.SendAndClose(&FileResponse{BytesWritten: totalBytesWritten})
				}
				return fmt.Errorf("error receiving input stream: %w", err)
			}

			once.Do(server.openOutputFile(outputFile, inputFile.Name, err))
			if err != nil {
				outputFile = nil
				return fmt.Errorf("error opening output file: %w", err)
			}

			var bytesWritten int64

			bytesWritten, err = io.Copy(outputFile, bytes.NewReader(inputFile.Content))
			if err != nil {
				return fmt.Errorf("error copying input stream contents: %w", err)
			}

			totalBytesWritten += bytesWritten
		}
	}
}

func (server *Server) openOutputFile(outputFile *os.File, fileName string, err error) func() {
	return func() {
		var openFile *os.File

		openFile, err = os.OpenFile(
			filepath.Clean(fmt.Sprintf("%s/%s", server.outputDirectory, fileName)),
			outputFileFlags,
			outputFilePermissions,
		)
		if err != nil {
			return
		}

		*outputFile = *openFile
	}
}
