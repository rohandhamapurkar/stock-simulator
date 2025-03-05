# Stock Market Simulator

A lightweight, Go-based simulation of a single-stock exchange with automatic order matching and price discovery.

## Overview

This project simulates a simplified stock exchange that trades a single stock. The simulation includes:

- Automatic generation of buy and sell orders
- Order matching engine based on price compatibility
- Real-time price updates when trades execute
- Binary Search Tree (BST) implementation for efficient order book management

The simulator runs indefinitely, continuously matching buy and sell orders and updating the Last Traded Price (LTP) when matches occur, providing a simplified view of how price discovery works in financial markets.

## Features

- **Single Stock Trading**: Focuses on the mechanics of order matching without the complexity of multiple securities
- **Automated Order Generation**: Creates random buy and sell orders based on the current market price
- **Price Discovery**: Demonstrates how market prices emerge from the interaction of buy and sell orders
- **Efficient Data Structures**: Uses Binary Search Trees for optimal order management
- **Concurrent Processing**: Leverages Go's goroutines for parallel processing of trade matching

## How It Works

The simulator consists of several key components:

1. **Exchange Engine**: Core component that maintains order books and matches trades
2. **Order Books**: Separate Binary Search Trees for buy and sell orders
3. **Order Generator**: Creates random buy and sell orders with prices around the current LTP
4. **Trade Processor**: Periodically checks for matching orders and executes trades

When buy and sell orders match (same price), a trade is executed, the orders are removed from their respective queues, and the Last Traded Price is updated to reflect the new market price.

## System Architecture

```
                 ┌─────────────────┐
                 │  Order Generator│
                 └────────┬────────┘
                          │
                          ▼
┌─────────┐      ┌─────────────────┐      ┌─────────────┐
│  Buy    │◄─────┤    Exchange     ├─────►│    Sell     │
│ Orders  │      │     Engine      │      │   Orders    │
└─────────┘      └────────┬────────┘      └─────────────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │ Trade Processor │
                 └─────────────────┘
```

## Usage

### Running the Simulator

Build and run the simulator:

```bash
go build
./stockmarketsim
```

The program will start generating random buy and sell orders, matching compatible orders, and updating the Last Traded Price. The current LTP will be displayed in the console as trades are executed.

To exit the simulator, press `Ctrl+C`.

### Understanding the Output

When the simulator is running, you'll see output like:

```
Processing trades
LTP: 95
Processing trades
LTP: 103
```

Each "LTP" line indicates a successful match between a buy and sell order, with the resulting price.

## Implementation Details

### Exchange

The exchange maintains two order books (implemented as Binary Search Trees) - one for buy orders and one for sell orders. It processes incoming orders and attempts to match them based on price compatibility.

### Transactions

Each transaction (order) includes:
- A unique identifier
- Type (BUY or SELL)
- Price amount

### Binary Search Tree

The order books use a Binary Search Tree data structure for efficient insertion, search, and deletion operations, which are critical for fast order matching.

### Concurrency

The simulator uses Go's goroutines and channels to handle concurrent operations:
- Order generation runs in a separate goroutine
- Order acceptance runs in its own goroutine
- Trade processing runs in another goroutine
- Mutex locks protect shared resources during updates

## Customization

You can modify the simulator's behavior by adjusting these parameters in `main.go`:

- Initial Last Traded Price (default: 100)
- Order generation frequency (default: 5 orders per second)
- Price range for buy orders (default: current price -100 to current price)
- Price range for sell orders (default: current price -25 to current price +100)

## License

```
MIT License

Copyright (c) 2025 Rohan Dhamapurkar

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Future Enhancements

- Add multiple stocks with different trading characteristics
- Implement more sophisticated order types (limit, market, stop)
- Add realistic market participants with different trading strategies
- Implement a visualization layer for real-time market data
- Add support for order cancellation and modification
- Implement trading volume statistics and market depth visualization
