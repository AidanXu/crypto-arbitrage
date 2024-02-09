package main

import (
	"context"
	"fmt"
	"log"
	"net"

	mycrypto "trade-service/protos"

	"google.golang.org/grpc"
)

type server struct {
    mycrypto.UnimplementedTradeStreamServer
}

// StreamTrades implements the StreamTrades method of the TradeSendServer interface
func (s *server) StreamTrades(ctx context.Context, req *mycrypto.TradeRequest) (*mycrypto.TradeResponse, error) {
    fmt.Printf("Received: %v\n", req)
    return &mycrypto.TradeResponse{}, nil
}

func main() {
    // Create a listener on TCP port
    lis, err := net.Listen("tcp", ":50052")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    // Create a gRPC server object
    s := grpc.NewServer()

    // Attach the TradeSend service to the server
    mycrypto.RegisterTradeStreamServer(s, &server{})

    // Serve gRPC server
    log.Println("Serving gRPC on 0.0.0.0:50052")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}