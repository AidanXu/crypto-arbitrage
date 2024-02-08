package main

import (
	"log"
	"net"

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
            S:    data.S,
            Bp:  data.Bp,
            Bs:   data.Bs,
            Ap:  data.Ap,
            As:   data.As,
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
    
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}