# Algorithmic Cryptocurrency Trading

This project is a microservice based algorithmic cryptocurrency trading platform using Go, gRPC, and Docker. It uses data from Binance to simulate scenarios and identify potential trading opportunities in the cryptocurrency market.

Currently the websocket stream and trading services are using Binance's spot trading testnet, but can be configured to work with the regular spot trading endpoints.

## Implemented Strategies

### Triangular Arbitrage

The core of my implemented strategies is the **triangular arbitrage method**, which capitalizes on discrepancies in exchange rates between three different cryptocurrencies on the same exchange. This method involves creating a cycle that starts and ends with the same currency, going through a series of trades that exploit price differences to secure a profit.

#### How It Works:

1. **Detection**: The platform continuously monitors the exchange rates between pairs of cryptocurrencies on Binance, using real-time data from the websocket stream.
2. **Opportunity Identification**: Employing a modified version of the **Shortest Path Faster Algorithm (SPFA)** with enhancements for negative cycle detection, the system efficiently identifies potential arbitrage opportunities. This algorithm excels at finding the most profitable paths for triangular arbitrage by detecting negative cycles in the graph of currency exchange rates.
3. **Execution**: Once a profitable triangular arbitrage path is identified, the system executes the series of trades almost simultaneously. This rapid execution is crucial to capitalize on the opportunity before the market adjusts the prices and the arbitrage opportunity disappears.

![Graph Example](https://thealgoristsblob.blob.core.windows.net/thealgoristsimages/arbitrage-2.png)

#### Advantages:

- **Low Risk and High Efficiency**: Triangular arbitrage within a single exchange eliminates transfer time and fees, making it a lower risk and more efficient strategy compared to inter-exchange arbitrage.
- **Automated Profit Generation**: By automating the detection and execution process, the system can swiftly react to arbitrage opportunities, generating profits with minimal manual intervention.
- **Scalability**: This strategy is part of a microservice-based architecture, ensuring the system remains scalable and adaptable to new strategies and improvements.

My current implementation uses Binance's spot trading testnet to validate the effectiveness of this strategy. The platform is designed with flexibility in mind, allowing for easy integration with other exchanges or trading environments.

## Currently Implementing

Right now I am working on managing a local order book using Binance's diff depth stream endpoint to implement order book imbalance strategies. Also I'm planning to stream the real-time data updates to the trading service directly to aviod having to call the rest/json api for last second price checks.

## Future Strategies

While the current implementation only supports triangular arbitrage, the project is designed to be extensible. More strategies will be added in the future to explore different algorithmic trading techniques.

## Deployment

To deploy this project install docker-compose, then run

```bash
  docker-compose up --build
```

in the root directory of the project.
