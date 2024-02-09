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
    // Create a new graph instance for the snapshot
    snapshot := New()

    // Lock the original graph to ensure consistency during the copy
    g.Mu.RLock() // Assuming you have a mutex `mu` in your Graph struct for synchronization
    defer g.Mu.RUnlock()

    // Iterate over the original graph data and copy it to the snapshot
    for from, edges := range g.data {
        if _, exists := snapshot.data[from]; !exists {
            snapshot.data[from] = make(map[string]Edge)
        }
        for to, edge := range edges {
            // Copy each edge. Since Edge struct contains only primitive types,
            // a simple assignment is enough for a deep copy here.
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
func (G *Graph) SPFA() (bool, []string) {
	dis := make(map[string]float64)
	pathLen := make(map[string]int)
    pre := make(map[string]string)
	queue := list.New()

	// Initialize distances and enqueue all vertices
	for v := range G.data {
		dis[v] = 0
		pathLen[v] = 0
        pre[v] = ""
		queue.PushBack(v)
	}

	// Number of vertices
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
                    tracedPath, found := Trace(pre, v)
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

	return false, []string{"nocycle"}
}

// contains checks if a value is in the queue.
func contains(queue *list.List, value string) bool {
	for e := queue.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == value {
			return true
		}
	}
	return false
}

func Trace(pre map[string]string, startV string) ([]string, bool) {
    visited := make(map[string]bool) // Tracks visited vertices to detect the cycle start
    var cycle []string // Stores the reconstructed cycle

    // Use 'v' to trace through the cycle, starting from 'startV'
    v := startV

    // 'prev' will track the last vertex we added to 'cycle' to prevent consecutive duplicates
    var prev string

    for {
        // Detect the cycle's start (where we loop back)
        if visited[v] {
            // Loop to construct the cycle from 'cycleStart' back to itself
            cycleStart := v
            tempCycle := []string{cycleStart} // Temporary slice to construct the cycle
            
            // Trace back the cycle, ensuring no consecutive vertices are identical
            for v = pre[cycleStart]; v != cycleStart; v = pre[v] {
                // Insert 'v' into 'tempCycle' if it's not identical to the last inserted vertex
                if v != prev {
                    tempCycle = append([]string{v}, tempCycle...)
                    prev = v
                } else {
                    // Found consecutive identical vertices, indicating an invalid cycle
                    return nil, false
                }
            }

            // Verify the cycle is valid (contains at least 3 distinct vertices)
            if len(tempCycle) >= 3 {
                // The cycle is valid; prepend 'cycleStart' to close the cycle
                return append([]string{cycleStart}, tempCycle...), true
            } else {
                return nil, false // The cycle does not meet the criteria
            }
        }

        // Mark the vertex as visited and add it to the cycle
        visited[v] = true
        if v != prev {
            cycle = append(cycle, v)
            prev = v
        }

        // Move to the predecessor, ensuring it exists
        nextV, exists := pre[v]
        if !exists {
            // The predecessor does not exist; invalid cycle
            return nil, false
        }
        v = nextV
    }
}