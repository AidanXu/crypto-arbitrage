package main

import (
	"fmt"
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

    // Process quotes as before
    for {
        data, err := stream.Recv()
        if err != nil {
            log.Printf("Error receiving data: %v", err)
            return err
        }

        quote := cryptoGraph.Quote{
            Symbol:    data.S,
            BidPrice:  data.Bp,
            BidSize:   data.Bs,
            AskPrice:  data.Ap,
            AskSize:   data.As,
            Timestamp: data.Time,
        }
        s.graph.AddQuote(quote)

        //fmt.Printf("Server: %v\n", data)
    }
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    g := cryptoGraph.New()

    s := grpc.NewServer()
    mycrypto.RegisterCryptoStreamServer(s, &server{graph: g})

    log.Println("Server is running on port 50051...")

    // Start the goroutine before serving
    go func(g *cryptoGraph.Graph) {
        for {
            arbitrageOpportunities := g.FindArbitrage()

            // If there are arbitrage opportunities, print them
            if len(arbitrageOpportunities) > 0 {
                fmt.Println("Arbitrage opportunities found:")
                for _, opportunity := range arbitrageOpportunities {
                    fmt.Println(opportunity)
                }
            }

            // Sleep for a while before the next run
            time.Sleep(time.Second/1000)
        }
    }(g)
    
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}