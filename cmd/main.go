package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"stockmarketsim/exchange"
	"stockmarketsim/ui"
	"syscall"
	"time"
)

func main() {
	// As of Go 1.20, rand.Seed is deprecated and no longer needed
	// The default global random source is automatically seeded with a random value

	// Create a logger for the main component
	logger := exchange.NewLogger("Main")
	logger.Info("Starting Stock Market Simulator")

	var ltp exchange.TransactionAmtDataType = 100
	logger.Info("Initializing exchange with LTP: 100")
	stockExchange := exchange.NewExchange(ltp)

	// Start the trade processing goroutine
	go stockExchange.ProcessTrades()

	// Start the trade acceptance goroutine
	go stockExchange.AcceptTrades()

	// Start the random trade generation goroutine
	go func() {
		generateRandomTrades(&stockExchange, logger)
	}()

	// Start the UI server
	logger.Info("Starting UI server")
	uiServer := ui.NewServer(&stockExchange)

	// Register a callback to broadcast price updates to UI clients
	stockExchange.RegisterPriceUpdateCallback(func(price int) {
		uiServer.BroadcastPriceUpdate(price)
	})

	// Start the UI server on port 8080
	uiServer.Start("8080")
	logger.Info("UI server started on http://localhost:8080")

	logger.Info("All systems initialized. Simulator running.")
	blockUntilSigInt(logger)
}

// generateRandomTrades generates random buy and sell orders at regular intervals
func generateRandomTrades(stkExch *exchange.Exchange, logger *exchange.Logger) {
	logger.Info("Starting random trade generation")
	ticker := time.NewTicker(time.Second)

	for {
		<-ticker.C
		currentPrice := int(stkExch.LastTradedPrice)

		for i := 0; i < 5; i++ {
			// Generate buy order
			buyPrice := getRandomIntForBuy(currentPrice)
			buyTxn := exchange.NewTransaction(
				exchange.BuyTransactionType,
				exchange.TransactionAmtDataType(buyPrice),
			)
			stkExch.IncomingTrades <- buyTxn
			logger.Debug("Generated buy order with price: " + fmt.Sprintf("%d", buyPrice))

			// Generate sell order
			sellPrice := getRandomIntForSell(currentPrice)
			sellTxn := exchange.NewTransaction(
				exchange.SellTransactionType,
				exchange.TransactionAmtDataType(sellPrice),
			)
			stkExch.IncomingTrades <- sellTxn
			logger.Debug("Generated sell order with price: " + fmt.Sprintf("%d", sellPrice))
		}
	}
}

// getRandomIntForBuy generates a random price for a buy order
// Ensures the price is at least 1 (minimum valid price)
func getRandomIntForBuy(target int) int {
	// Set minimum price to max(1, target-100)
	min := max(1, target-100)

	// Set maximum price to max(target, min+1)
	maxPrice := max(target, min+1)

	return rand.Intn(maxPrice-min+1) + min
}

// getRandomIntForSell generates a random price for a sell order
// Ensures the price is at least 1 (minimum valid price)
func getRandomIntForSell(target int) int {
	// Set minimum price to max(1, target-25)
	min := max(1, target-25)

	// Set maximum price to max(target+100, min+1)
	maxPrice := max(target+100, min+1)

	return rand.Intn(maxPrice-min+1) + min
}

// blockUntilSigInt blocks until a SIGINT (Ctrl+C) is received
func blockUntilSigInt(logger *exchange.Logger) {
	// Create a channel to receive OS signals
	signalChannel := make(chan os.Signal, 1)

	// Notify the signal channel on receiving SIGINT (Ctrl+C) or SIGTERM signals
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Press Ctrl+C to exit...")

	// Block until a signal is received
	<-signalChannel

	// Handle the received signal
	logger.Info("Received shutdown signal. Exiting...")
}
