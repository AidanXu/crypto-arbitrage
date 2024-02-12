package binance

import (
	"crypto/sha1"
	"encoding/hex"
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

func GenerateRouteHash(route []*mycrypto.TradeInfo) string {
    var hashInput string
    for _, trade := range route {
        hashInput += fmt.Sprintf("%s-%s-%f-%f|", trade.S, trade.E, trade.Rate, trade.Size)
    }
    hash := sha1.New()
    hash.Write([]byte(hashInput))
    return hex.EncodeToString(hash.Sum(nil))
}

var checkedRoutes = make(map[string]bool)

func CheckAndStoreRoute(route []*mycrypto.TradeInfo) bool {
    hash := GenerateRouteHash(route)
    if _, exists := checkedRoutes[hash]; exists {
        // Route has already been checked
        return false
    }
    // Mark the route as checked
    checkedRoutes[hash] = true
    return true
}

func CheckRoute(tradeRoute []*mycrypto.TradeInfo) () {

	fmt.Printf("Trade Route: %v\n", tradeRoute)

	if !CheckAndStoreRoute(tradeRoute) {
		// Route has already been checked
		return
	}

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
    fmt.Printf("Current Prices: %s\n", response)

    return
}

func convertTradeRouteToSymbols(tradeRoute []*mycrypto.TradeInfo) []string {
    // Define the valid symbols
    validSymbols := map[string]bool{
        "BTCUSDT": true,
        "ETHUSDT": true,
        "BNBUSDT": true,
        "XRPUSDT": true,
        "ADAUSDT": true,
        "SOLUSDT": true,
        "DOTUSDT": true,
        "LTCUSDT": true,
        "BCHUSDT": true,
        "LINKUSDT": true,
        "XLMUSDT": true,
        "UNIUSDT": true,
        "DOGEUSDT": true,
        "WBTCUSDT": true,
        "AAVEUSDT": true,
        "ATOMUSDT": true,
        "ETHBTC": true,
        "BNBBTC": true,
        "BNBETH": true,
        "XRPBTC": true,
        "XRPETH": true,
        "ADABTC": true,
        "ADAETH": true,
        "SOLBTC": true,
        "SOLETH": true,
        "DOTBTC": true,
        "DOTETH": true,
        "LTCBTC": true,
        "LTCETH": true,
        "BCHBTC": true,
        "LINKBTC": true,
        "LINKETH": true,
        "XLMBTC": true,
        "XLMETH": true,
        "UNIBTC": true,
        "UNIETH": true,
        "DOGEBTC": true,
        "DOGEETH": true,
        "WBTCETH": true,
        "AAVEBTC": true,
        "AAVEETH": true,
        "ATOMBTC": true,
        "ATOMETH": true,
    }

    symbols := make([]string, 0, len(tradeRoute))
    for _, trade := range tradeRoute {

        symbol1 := fmt.Sprintf("%s%s", trade.S, trade.E)
        symbol2 := fmt.Sprintf("%s%s", trade.E, trade.S)

        // Check if the first symbol is valid
        if _, ok := validSymbols[symbol1]; ok {
            symbols = append(symbols, symbol1)
        } else if _, ok := validSymbols[symbol2]; ok {
            symbols = append(symbols, symbol2)
        }
    }
    return symbols
}