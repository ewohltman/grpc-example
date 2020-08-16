package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/ewohltman/grpc-example/pkg/filetransfer"
)

const serverAddress = "localhost:8080"

func main() {
	log := logrus.NewEntry(logrus.New())
	log.Logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	if len(os.Args) <= 1 {
		log.Fatal("File name must be provided")
	}

	log.Info("Starting up grpc-client")

	client, err := filetransfer.NewClient(log, serverAddress)
	if err != nil {
		log.WithError(err).Fatal("Error starting grpc-client")
	}

	defer func() {
		err := client.Close()
		if err != nil {
			log.WithError(err).Errorf("Error closing connection")
		}
	}()

	client.UploadReader(context.TODO(), os.Args[1], os.Stdin)
}
