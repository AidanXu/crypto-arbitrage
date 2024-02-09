package main

import (
	"log"
	"net"
	"sync"
	"time"

	cryptoGraph "detection-algos/graph"
	mycrypto "detection-algos/protos"

	"google.golang.org/grpc"
)

type server struct {
    mycrypto.UnimplementedCryptoStreamServer
    graph *cryptoGraph.Graph
    mu    sync.RWMutex
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
        s.mu.Lock();
        s.graph.AddQuote(quote);
        s.mu.Unlock();
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

        for range ticker.C {
            srv.mu.RLock()
            arbitragePaths := srv.graph.DetectNegativeCycle()
            srv.mu.RUnlock()
            if (arbitragePaths) == true {
                //log.Println("Arbitrage opportunity detected:")
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
