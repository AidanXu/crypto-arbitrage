package main

import (
	"context"
	"log"
	"net"
	"time"

	cryptoGraph "detection-algos/graph"
	mycrypto "detection-algos/protos"

	"google.golang.org/grpc"
)

type server struct {
    mycrypto.UnimplementedCryptoStreamServer
    graph *cryptoGraph.Graph
}

func (s *server) StreamCrypto(stream mycrypto.CryptoStream_StreamCryptoServer) error {

    for {
        data, err := stream.Recv()
        if err != nil {
            log.Printf("Error receiving data: %v", err)
            return err
        }

        quote := cryptoGraph.Quote{
            S:    data.S,
            Bp:  data.Bp,
            Bs:   data.Bs,
            Ap:  data.Ap,
            As:   data.As,
        }
        s.graph.Mu.Lock();
        s.graph.AddQuote(quote);
        s.graph.Mu.Unlock();
    }
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    g := cryptoGraph.New()

    srv := &server{graph: g}

    s := grpc.NewServer()
    mycrypto.RegisterCryptoStreamServer(s, srv)

    //Goroutine for periodic arbitrage checks
    go func(srv *server) {
        ticker := time.NewTicker(time.Millisecond)
        defer ticker.Stop()

        conn, err := grpc.Dial("trade-service:50052", grpc.WithInsecure())
        if err != nil {
            log.Printf("Failed to connect: %v", err)
        }
        defer conn.Close()

        client := mycrypto.NewTradeStreamClient(conn)

        for range ticker.C {
            snapshot := srv.graph.Snapshot()
            found, arbitragePaths := snapshot.SPFA()
            if found {

                tradeRequest := &mycrypto.TradeRequest{
                    TradeRoute: make([]*mycrypto.TradeInfo, len(arbitragePaths.Route)),
                }
                for i, step := range arbitragePaths.Route {
                    tradeRequest.TradeRoute[i] = &mycrypto.TradeInfo{
                        S:     step.From,
                        E:     step.To,
                        Rate:  float32(step.EdgeData.Rate),
                        Size:  float32(step.EdgeData.Size),
                    }
                }
        
                if _, err := client.StreamTrades(context.Background(), tradeRequest); err != nil {
                    log.Fatalf("Failed to send trade request: %v", err)
                }
            } else {
                //log.Println("No arbitrage opportunities detected")
            }
        }
    }(srv)

    log.Println("Server is running on port 50051...")
    
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
