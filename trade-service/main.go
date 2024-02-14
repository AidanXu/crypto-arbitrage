package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	binance "trade-service/binance"
	mycrypto "trade-service/protos"

	"google.golang.org/grpc"
)

type server struct {
    mycrypto.UnimplementedTradeStreamServer
    mu sync.Mutex
    count int
    lowestSum float64
    lowestRoute []*mycrypto.TradeInfo
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
        s.lowestRoute = req.TradeRoute
    }

    s.count++
    s.mu.Unlock()

    // Filter for reasonably profitable routes
    if sum < -0.005 {
        binance.CheckRoute(req.TradeRoute)
    }

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
            lowestSum := tradeServer.lowestSum
            totalTrades := tradeServer.count
            //lowestRoute := tradeServer.lowestRoute
            tradeServer.count = 0
            tradeServer.lowestSum = 0
            tradeServer.lowestRoute = nil

            log.Printf("Lowest Sum: %f, Total Routes: %d", lowestSum, totalTrades)
            // if lowestRoute != nil {
            //     log.Printf("Route with lowest sum: %v", lowestRoute)
            // }
        }
    }()

    // Attach the TradeSend service to the server
    mycrypto.RegisterTradeStreamServer(grpcServer, tradeServer)

    // Serve gRPC server
    log.Println("Serving gRPC on 50052")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}