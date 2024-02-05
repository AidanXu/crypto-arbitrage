package cryptoGraph

import (
	"log"
	"math"
	"strings"
)

type Quote struct {
    Symbol    string
    BidPrice  float64
    BidSize   float64
    AskPrice  float64
    AskSize   float64
    Timestamp string
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
	// don't add logging
    if !strings.Contains(quote.Symbol, "/") {
        log.Printf("Invalid symbol: %s", quote.Symbol)
        return
    }

    currencies := strings.Split(quote.Symbol, "/")
    base, quoteCurrency := currencies[0], currencies[1]

    // Update edges
    if _, exists := (g.data)[base]; !exists {
        (g.data)[base] = make(map[string]Edge)
    }
    (g.data)[base][quoteCurrency] = Edge{Rate: 1 / quote.BidPrice, Size: quote.BidSize}

    if _, exists := (g.data)[quoteCurrency]; !exists {
        (g.data)[quoteCurrency] = make(map[string]Edge)
    }
    (g.data)[quoteCurrency][base] = Edge{Rate: quote.AskPrice, Size: quote.AskSize}

	//fmt.Printf("Edge created: %s -> %s with Rate: %f and Size: %f\n", base, quoteCurrency, 1/quote.BidPrice, quote.BidSize)
    //fmt.Printf("Edge created: %s -> %s with Rate: %f and Size: %f\n", quoteCurrency, base, quote.AskPrice, quote.AskSize)
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
