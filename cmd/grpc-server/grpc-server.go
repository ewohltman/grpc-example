package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/ewohltman/grpc-example/pkg/filetransfer"
	"github.com/ewohltman/grpc-example/pkg/grpcServer"
)

const port = 8080

func main() {
	log.Println("Starting up grpc-server")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %s\n", err)
	}

	server := grpc.NewServer()

	filetransfer.RegisterFileTransferServer(server, &grpcServer.Server{})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	go func() {
		err := server.Serve(listener)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	log.Println("Start up grpc-server complete. Listening...")

	<-stop

	server.GracefulStop()
}
