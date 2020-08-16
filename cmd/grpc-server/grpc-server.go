package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/ewohltman/grpc-example/pkg/filetransfer"
)

const (
	network         = "tcp"
	address         = ":8080"
	outputDirectory = "/tmp"
)

func main() {
	log := logrus.NewEntry(logrus.New())
	log.Logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	log.Info("Starting up grpc-server")

	serverConfig := &filetransfer.ServerConfig{
		Log:             log,
		Network:         network,
		Address:         address,
		OutputDirectory: outputDirectory,
	}

	gracefulStop, err := filetransfer.NewServer(serverConfig)
	if err != nil {
		log.Fatal(err)
	}

	defer gracefulStop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	log.Info("Server started up. Listening...")

	<-stop
}
