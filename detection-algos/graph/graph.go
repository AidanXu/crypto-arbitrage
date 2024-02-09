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
       if edge, exists := baseEdges[quoteCurrency]; exists {
           if edge.Rate != transformedBidRate {
               // If the rate has changed, update it
               (g.data)[base][quoteCurrency] = Edge{Rate: transformedBidRate, Size: quote.Bs}
               //log.Printf("Updated edge: %s -> %s: %+v", base, quoteCurrency, (g.data)[base][quoteCurrency])
           }
       } else {
           // If the edge doesn't exist, create it
           (g.data)[base][quoteCurrency] = Edge{Rate: transformedBidRate, Size: quote.Bs}
           //log.Printf("Created edge: %s -> %s: %+v", base, quoteCurrency, (g.data)[base][quoteCurrency])
       }
   } else {
       (g.data)[base] = make(map[string]Edge)
       (g.data)[base][quoteCurrency] = Edge{Rate: transformedBidRate, Size: quote.Bs}
       //log.Printf("Created edge: %s -> %s: %+v", base, quoteCurrency, (g.data)[base][quoteCurrency])
   }

    quoteEdges, quoteExists := (g.data)[quoteCurrency]
    if quoteExists {
        if edge, exists := quoteEdges[base]; exists {
            if edge.Rate != transformedAskRate {
                // If the rate has changed, update it
                (g.data)[quoteCurrency][base] = Edge{Rate: transformedAskRate, Size: quote.As}
                //log.Printf("Updated edge: %s -> %s: %+v", quoteCurrency, base, (g.data)[quoteCurrency][base])
            }
        } else {
            // If the edge doesn't exist, create it
            (g.data)[quoteCurrency][base] = Edge{Rate: transformedAskRate, Size: quote.As}
            //log.Printf("Created edge: %s -> %s: %+v", quoteCurrency, base, (g.data)[quoteCurrency][base])
        }
    } else {
        (g.data)[quoteCurrency] = make(map[string]Edge)
        (g.data)[quoteCurrency][base] = Edge{Rate: transformedAskRate, Size: quote.As}
        //log.Printf("Created edge: %s -> %s: %+v", quoteCurrency, base, (g.data)[quoteCurrency][base])
    }

}


// Using bellman ford
func (g *Graph) DetectNegativeCycle() bool {
    // Initialization
    dist := make(map[string]float64)
    for vertex := range g.data {
        dist[vertex] = math.MaxFloat64
    }

    var startVertex string
    for vertex := range g.data {
        startVertex = vertex
        break
    }
    dist[startVertex] = 0

    // Relaxation
    for i := 0; i < len(g.data)-1; i++ {
        updateOccurred := false
        for u, neighbors := range g.data {
            for v, edge := range neighbors {
                if dist[u] != math.MaxFloat64 && dist[u]+edge.Rate < dist[v] {
                    dist[v] = dist[u] + edge.Rate
                    updateOccurred = true
                }
            }
        }
        if !updateOccurred {
            break
        }
    }

    // Check for negative cycles
    for u, neighbors := range g.data {
        for v, edge := range neighbors {
            if dist[u] != math.MaxFloat64 && dist[u]+edge.Rate < dist[v] {
                return true // Negative cycle detected
            }
        }
    }

    return false // No negative cycle found
}

// Using shortest path faster for negative cycle detection and reconstruction
