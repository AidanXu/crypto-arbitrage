package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	mycrypto "trade-service/protos"
)

// Response struct
type SymbolPriceTicker struct {
    Symbol string `json:"symbol"`
    Price  string `json:"price"`
}

func CheckRoute(tradeRoute []*mycrypto.TradeInfo) () {

    symbols := convertTradeRouteToSymbols(tradeRoute)
    
    // Remove spaces
    formattedSymbols := make([]string, len(symbols))
    for i, symbol := range symbols {
        symbol = strings.ReplaceAll(symbol, " ", "")
        formattedSymbols[i] = symbol
    }

    // Convert formattedSymbols slice to JSON string
    symbolsJSON, err := json.Marshal(formattedSymbols)
    if err != nil {
        fmt.Printf("Error marshaling symbols to JSON: %v\n", err)
        return
    }

    fmt.Printf("Symbols JSON: %s\n", symbolsJSON)

    // Prepare params for the API call
    params := map[string]string{
        "symbols": string(symbolsJSON),
    }

    client := NewClient();

    // Get current market price
    response, err := client.DoGetRequest("/api/v3/ticker/price", params)
    if err != nil {
        log.Fatalf("Failed to do GET request: %v", err)
    }

	// Calculate if still profitable and make trade if so using limit orders
    fmt.Printf("Response: %s\n", response)

    return
}

func convertTradeRouteToSymbols(tradeRoute []*mycrypto.TradeInfo) []string {
    symbols := make([]string, 0, len(tradeRoute))
    for _, trade := range tradeRoute {
        // Check if the symbol is one of the exceptions
        if (trade.S == "BTC" && trade.E == "USDT") || 
           (trade.S == "ETH" && trade.E == "USDT") || 
           (trade.S == "ETH" && trade.E == "BTC") {
            symbol := fmt.Sprintf("%s%s", trade.S, trade.E)
            symbols = append(symbols, symbol)
        } else if trade.E == "BTC" || trade.E == "USDT" || trade.E == "ETH" {
            // If not, check if the second parameter is "BTC", "USDT", or "ETH"
            symbol := fmt.Sprintf("%s%s", trade.S, trade.E)
            symbols = append(symbols, symbol)
        } else {
            // If not, swap the order of the parameters
            symbol := fmt.Sprintf("%s%s", trade.E, trade.S)
            symbols = append(symbols, symbol)
        }
    }
    return symbols
}