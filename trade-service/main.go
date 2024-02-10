package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	mycrypto "trade-service/protos"

	"google.golang.org/grpc"
)

type server struct {
    mycrypto.UnimplementedTradeStreamServer
    mu sync.Mutex
    sum float64
    count int
    lowestSum float64
}

// StreamTrades implements the StreamTrades method of the TradeSendServer interface
func (s *server) StreamTrades(ctx context.Context, req *mycrypto.TradeRequest) (*mycrypto.TradeResponse, error) {
    sum := 0.0
    for _, trade := range req.TradeRoute {
        sum += float64(trade.Rate)
    }

    s.mu.Lock()
    if s.count == 0 || sum < s.lowestSum {
        s.lowestSum = sum
    }
    s.sum += sum
    s.count++
    s.mu.Unlock()

    return &mycrypto.TradeResponse{}, nil
}

func main() {
    // Create a listener on TCP port
    lis, err := net.Listen("tcp", ":50052")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    // Create a gRPC server object
    grpcServer := grpc.NewServer()

    // Create a new server
    tradeServer := &server{}

    // Start a goroutine to calculate the average every 10 seconds
    go func() {
        for range time.Tick(10 * time.Second) {
            tradeServer.mu.Lock()
            average := tradeServer.sum / float64(tradeServer.count)
            lowestSum := tradeServer.lowestSum
            tradeServer.sum = 0
            tradeServer.count = 0
            tradeServer.lowestSum = 0
            tradeServer.mu.Unlock()

            log.Printf("Average sum of loop: %f, Lowest Sum: %f", average, lowestSum)
        }
    }()

    // Attach the TradeSend service to the server
    mycrypto.RegisterTradeStreamServer(grpcServer, tradeServer)

    // Serve gRPC server
    log.Println("Serving gRPC on 0.0.0.0:50052")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}