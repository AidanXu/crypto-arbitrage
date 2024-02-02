package main

import (
	"fmt"
	"log"
	"net"

	mycrypto "detection-algos/protos"

	"google.golang.org/grpc"
)

type server struct {
    mycrypto.UnimplementedCryptoStreamServer
}

func (s *server) StreamCrypto(stream mycrypto.CryptoStream_StreamCryptoServer) error {
    for {
        data, err := stream.Recv()
        if err != nil {
            log.Printf("Error receiving data: %v", err)
            return err
        }

        fmt.Printf("Server: %v\n", data)
    }
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    s := grpc.NewServer()
    mycrypto.RegisterCryptoStreamServer(s, &server{})

    log.Println("Server is running on port 50051...")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}