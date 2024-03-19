package binance

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	mycrypto "trade-service/protos"
)

func GenerateRouteHash(route []*mycrypto.TradeInfo) string {
    var hashInput string
    for _, trade := range route {
        hashInput += fmt.Sprintf("%s-%s|", trade.S, trade.E)
    }
    hash := sha1.New()
    hash.Write([]byte(hashInput))
    return hex.EncodeToString(hash.Sum(nil))
}

var (
    checkedRoutes map[string]bool
    mapMutex      sync.Mutex
)

func init() {
    checkedRoutes = make(map[string]bool)
    go periodicallyClearMap()
}

// Reset the checkedRoutes map every minute
func periodicallyClearMap() {
    ticker := time.NewTicker(3*time.Minute) 
    for range ticker.C {
        mapMutex.Lock()
        checkedRoutes = make(map[string]bool)
        mapMutex.Unlock()
        fmt.Printf("checkRoutes map cleared \n")
    }
}

func CheckAndStoreRoute(route []*mycrypto.TradeInfo) bool {
    hash := GenerateRouteHash(route)
	mapMutex.Lock()
    if _, exists := checkedRoutes[hash]; exists {
        mapMutex.Unlock()
        return false
    }
    checkedRoutes[hash] = true
	mapMutex.Unlock()
    return true
}

func CheckRoute(tradeRoute []*mycrypto.TradeInfo) () {

	if !CheckAndStoreRoute(tradeRoute) {
		return
	}

	fmt.Printf("Trade Route: %v\n", tradeRoute)

    // Get current price of symbols
    symbols := convertTradeRouteToSymbols(tradeRoute)
    
    formattedSymbols := make([]string, len(symbols))
    for i, symbol := range symbols {
        symbol = strings.ReplaceAll(symbol, " ", "")
        formattedSymbols[i] = symbol
    }

    symbolsJSON, err := json.Marshal(formattedSymbols)
    if err != nil {
        fmt.Printf("Error marshaling symbols to JSON: %v\n", err)
        return
    }

    params := map[string]string{
        "symbols": string(symbolsJSON),
    }

    client := NewClient();

    response, err := client.DoGetRequest("/api/v3/ticker/price", params)
    if err != nil {
        log.Printf("Failed to do GET request: %v", err)
    }

    var prices []SymbolPriceTicker
    err = json.Unmarshal([]byte(response), &prices)
    if err != nil {
        log.Printf("Failed to unmarshal response: %v", err)
    }

    fmt.Printf("Current Prices: %+v\n", prices)

    recalculated_sum := 0.0

    for _, trade := range tradeRoute {
        currentRate := getCurrentRates(trade.S, trade.E, prices)
        recalculated_sum += -math.Log(currentRate)
    }

    min_return := -0.001 * float64(len(tradeRoute))

    if (recalculated_sum < min_return) {
        fmt.Printf("Sum of path with current rates: %v\n", recalculated_sum)
        // TODO: implement place trade
        // placeTrade(tradeRoute)
    } else {
        fmt.Printf("Route no longer profitable\n")
    }

    return
}

// Response struct
type SymbolPriceTicker struct {
    Symbol string `json:"symbol"`
    Price  string `json:"price"`
}

func getCurrentRates(start string, end string, current_prices []SymbolPriceTicker) float64 {
    symbol := start + end
    inverseSymbol := end + start

    for _, price := range current_prices {
        if price.Symbol == symbol {
            parsedPrice, err := strconv.ParseFloat(price.Price, 64)
            if err != nil {
                log.Printf("Failed to parse price: %v", err)
            }
            return parsedPrice
        } else if price.Symbol == inverseSymbol {
            parsedPrice, err := strconv.ParseFloat(price.Price, 64)
            if err != nil {
                log.Printf("Failed to parse price: %v", err)
            }
            return 1 / parsedPrice
        }
    }
    return 0
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