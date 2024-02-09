package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"

	mycrypto "datacollection/protos"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

func main() {
    u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws"}

    c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("dial:", err)
    }
    defer c.Close()

    subscribe := map[string]interface{}{
        "method": "SUBSCRIBE",
        "params": []string{
			"btcusdt@bookTicker",
			"ethusdt@bookTicker",
			// "bnbusdt@bookTicker",
			// "xrpusdt@bookTicker",
			// "adausdt@bookTicker",
			// "solusdt@bookTicker",
			// "dotusdt@bookTicker",
			// "ltcusdt@bookTicker",
			// "bchusdt@bookTicker",
			// "linkusdt@bookTicker",
			// "xlmusdt@bookTicker",
			// "uniusdt@bookTicker",
			// "dogeusdt@bookTicker",
			// "wbtcusdt@bookTicker",
			// "aaveusdt@bookTicker",
			// "atomusdt@bookTicker",
			"ethbtc@bookTicker",
			// "bnbbtc@bookTicker",
			// "bnbeth@bookTicker",
			// "xrpbtc@bookTicker",
			// "xrpeth@bookTicker",
			// "adabtc@bookTicker",
			// "adaeth@bookTicker",
			// "solbtc@bookTicker",
			// "soleth@bookTicker",
			// "dotbtc@bookTicker",
			// "doteth@bookTicker",
			// "ltcbtc@bookTicker",
			// "ltceth@bookTicker",
			// "bchbtc@bookTicker",
			// "linkbtc@bookTicker",
			// "linketh@bookTicker",
			// "xlmbtc@bookTicker",
			// "xlmeth@bookTicker",
			// "unibtc@bookTicker",
			// "unieth@bookTicker",
			// "dogebtc@bookTicker",
			// "dogeeth@bookTicker",
			// "wbtceth@bookTicker",
			// "aavebtc@bookTicker",
			// "aaveeth@bookTicker",
			// "atombtc@bookTicker",
			// "atometh@bookTicker",
        },
        "id": 1,
    }

    err = c.WriteJSON(subscribe)
    if err != nil {
        log.Fatal("subscribe:", err)
    }

	conn, err := grpc.Dial("detection-algos:50051", grpc.WithInsecure())
	if err != nil {
		log.Printf("did not connect: %v", err)
		// Don't return, just log the error
	}
	defer conn.Close()

	var stream mycrypto.CryptoStream_StreamCryptoClient
	if err == nil {
		client := mycrypto.NewCryptoStreamClient(conn)

		// Start the stream
		stream, err = client.StreamCrypto(context.Background())
		if err != nil {
			log.Printf("Error on stream: %v", err)
			// Don't return, just log the error
		}
	}

    for {
        _, message, err := c.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            return
        }

        //log.Printf("recv: %s", message)

		type CryptoDataJSON struct {
			S  string  `json:"s"`
			Bp float64 `json:"b,string"`
			Bs float64 `json:"B,string"`
			Ap float64 `json:"a,string"`
			As float64 `json:"A,string"`
		}

		var data CryptoDataJSON
		err = json.Unmarshal(message, &data)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		if data.S == "" {
			continue
		}

		cryptoData := &mycrypto.CryptoData{
			S:  data.S,
			Bp: data.Bp,
			Bs: data.Bs,
			Ap: data.Ap,
			As: data.As,
		}

		if err := stream.Send(cryptoData); err != nil {
			log.Fatalf("Failed to send data: %v", err)
		}

    }
}