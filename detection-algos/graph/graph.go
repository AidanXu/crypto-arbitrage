package cryptoGraph

import (
	"container/list"
	"log"
	"math"
	"regexp"
	"sync"
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

type RouteStep struct {
    From   string
    To     string
    EdgeData Edge
}

type RouteInfo struct {
    Route []RouteStep
}

type Graph struct {
    data map[string]map[string]Edge
    Mu  sync.RWMutex
}

func New() *Graph {
    return &Graph{
        data: make(map[string]map[string]Edge),
    }
}

func (g *Graph) Snapshot() *Graph {
    snapshot := New()

    g.Mu.RLock() 
    defer g.Mu.RUnlock()

    for from, edges := range g.data {
        if _, exists := snapshot.data[from]; !exists {
            snapshot.data[from] = make(map[string]Edge)
        }
        for to, edge := range edges {
            snapshot.data[from][to] = Edge{
                Rate: edge.Rate,
                Size: edge.Size,
            }
        }
    }

    return snapshot
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
    // Assuming a 0.1% transaction fee for simplicity
    transactionFee := 0.001

    effectiveBidRate := quote.Bp * (1 - 2*transactionFee)
    transformedBidRate := -math.Log(1 / effectiveBidRate)

    effectiveAskRate := quote.Ap * (1 + 2*transactionFee)
    transformedAskRate := -math.Log(effectiveAskRate)


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
func (G *Graph) SPFA() (bool, RouteInfo) {
	dis := make(map[string]float64)
	pathLen := make(map[string]int)
    pre := make(map[string]string)
	queue := list.New()

	for v := range G.data {
		dis[v] = 0
		pathLen[v] = 0
        pre[v] = ""
		queue.PushBack(v)
	}

	n := len(G.data)

	for queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		u := element.Value.(string)

		if _, exists := G.data[u]; !exists {
			continue
		}

		for v, edge := range G.data[u] {
			if dis[u]+edge.Rate < dis[v] {
				dis[v] = dis[u] + edge.Rate
                pre[v] = u
				pathLen[v] = pathLen[u] + 1
				if pathLen[v] >= n {
                    // Negative cycle detected
                    tracedPath, found := Trace(G, pre, v)
					if (found) {
                        return true, tracedPath
                    }
                    continue;
				}
				// Check if v is not already in the queue
				if !contains(queue, v) {
					queue.PushBack(v)
				}
			}
		}
	}
    var empty RouteInfo;
	return false, empty
}

func contains(queue *list.List, value string) bool {
	for e := queue.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == value {
			return true
		}
	}
	return false
}

func Trace(G *Graph, pre map[string]string, startV string) (RouteInfo, bool) {
    visited := make(map[string]bool)
    var tempCycle []string 
    
    v := startV 
    prev := ""

    for {
        if visited[v] {
            cycleStart := v
            tempCycle = append(tempCycle, cycleStart)

            var routeSteps []RouteStep
            isValidCycle := len(tempCycle) > 1 && tempCycle[0] == tempCycle[len(tempCycle)-1]

            // Iterate through tempCycle to construct routeSteps with edge data
            for i := 0; i < len(tempCycle)-1 && isValidCycle; i++ {
                from := tempCycle[i]
                to := tempCycle[i+1]
                if from == to { 
                    return RouteInfo{}, false
                }
                edge, exists := G.data[from][to]
                if !exists { // Edge must exist for a valid route step
                    return RouteInfo{}, false
                }
                routeSteps = append(routeSteps, RouteStep{From: from, To: to, EdgeData: Edge{Rate: edge.Rate, Size: edge.Size}})
            }

            // at least 3 distinct vertices
            if len(routeSteps) >= 3 && isValidCycle {
                return RouteInfo{Route: routeSteps}, true
            } else {
                return RouteInfo{}, false
            }
        }

        visited[v] = true
        tempCycle = append(tempCycle, v)

        nextV, exists := pre[v]
        if !exists || v == prev { 
            return RouteInfo{}, false
        }
        prev = v
        v = nextV
    }
}
