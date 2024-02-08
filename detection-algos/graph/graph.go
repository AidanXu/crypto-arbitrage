package cryptoGraph

import (
	"log"
	"math"
	"regexp"
)

type Quote struct {
    S    string
    Bp  float64
    Bs   float64
    Ap  float64
    As   float64
}

type Edge struct {
    Rate float64
    Size float64
}

type Graph struct {
    data map[string]map[string]Edge
}

func New() *Graph {
    return &Graph{
        data: make(map[string]map[string]Edge),
    }
}

func (g *Graph) AddQuote(quote Quote) {
    if quote.S == "" {
        log.Printf("Invalid symbol: %s", quote.S)
        return
    }

    re := regexp.MustCompile(`([A-Z]+)(BTC|ETH|BNB|XRP|ADA|SOL|DOT|LTC|BCH|LINK|XLM|UNI|DOGE|WBTC|AAVE|ATOM|USDT)`)
    match := re.FindStringSubmatch(quote.S)
    if len(match) < 3 {
        log.Printf("Invalid symbol: %s", quote.S)
        return
    }

    base, quoteCurrency := match[1], match[2]

    // Preprocessing: Using negative logarithm of the exchange rates
    transformedBidRate := -math.Log(1 / quote.Bp)
    transformedAskRate := -math.Log(quote.Ap)

    // Check for existing rates and only update if there's a change
    baseEdges, baseExists := (g.data)[base]
    if baseExists {
        if edge, exists := baseEdges[quoteCurrency]; exists && edge.Rate == transformedBidRate {
            // If the rate hasn't changed, return without updating
            return
        }
    } else {
        (g.data)[base] = make(map[string]Edge)
    }

    // Since there's a change or the pair doesn't exist, update the rate
    (g.data)[base][quoteCurrency] = Edge{Rate: transformedBidRate, Size: quote.Bs}

    quoteEdges, quoteExists := (g.data)[quoteCurrency]
    if quoteExists {
        if edge, exists := quoteEdges[base]; exists && edge.Rate == transformedAskRate {
            // If the rate hasn't changed, return without updating
            return
        }
    } else {
        (g.data)[quoteCurrency] = make(map[string]Edge)
    }

    // Since there's a change or the pair doesn't exist, update the rate
    (g.data)[quoteCurrency][base] = Edge{Rate: transformedAskRate, Size: quote.As}
}

func (g *Graph) FindArbitrage() [][]string {
    dist := make(map[string]float64)
    prev := make(map[string]string)
    for vertex := range g.data {
        dist[vertex] = math.MaxFloat64
    }

    var source string
    for vertex := range g.data {
        source = vertex
        break
    }
    dist[source] = 0

    // Relax edges |V| - 1 times
    for i := 0; i < len(g.data)-1; i++ {
        for vertex, edges := range g.data {
            for neighbor, edge := range edges {
                if dist[vertex]+-math.Log(edge.Rate) < dist[neighbor] {
                    dist[neighbor] = dist[vertex] + -math.Log(edge.Rate)
                    prev[neighbor] = vertex
                }
            }
        }
    }

    // Check for a negative-weight cycle
    arbitrageCycles := make([][]string, 0)
    for vertex, edges := range g.data {
        for neighbor, edge := range edges {
            if dist[vertex]+-math.Log(edge.Rate) < dist[neighbor] {
                // Negative cycle found, trace back the path
                cycle := []string{neighbor}
                for v := vertex; v != neighbor; v = prev[v] {
                    cycle = append([]string{v}, cycle...)
                }
                cycle = append([]string{neighbor}, cycle...)
                arbitrageCycles = append(arbitrageCycles, cycle)
            }
        }
    }

    return arbitrageCycles
}
