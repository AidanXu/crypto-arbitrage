# Crypto Arbitrage Trading Simulation

This project is a microservice-based cryptocurrency arbitrage trading simulation using Go, gRPC, and Docker. It uses data from Binance to simulate trading scenarios and identify potential arbitrage opportunities in the cryptocurrency market.

Currently the websocket stream and trading services are using Binance's spot trading testnet, but can be configured to work with the regular spot trading API.

## Implemented Strategy

The current implementation focuses on triangular arbitrage using a modified version of the shortest path faster algorithm with negative cycle tracing and recreation. This strategy allows the system to identify and exploit arbitrage opportunities by finding the most profitable triangular trading paths.

## Future Strategies

While the current implementation only supports triangular arbitrage, the project is designed to be extensible. More strategies will be added in the future to explore different algorithmic trading techniques.
