package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"stockmarketsim/exchange"
	"syscall"
	"time"
)

func main() {

	var ltp exchange.TransactionAmtDataType = 100
	stockExchange := exchange.NewExchange(ltp)

	go stockExchange.ProcessTrades()
	go stockExchange.AcceptTrades()

	go func() {
		generateRandomTrades(&stockExchange)
	}()

	blockUntilSigInt()
}

func generateRandomTrades(stkExch *exchange.Exchange) {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		for i := 0; i < 5; i++ {
			stkExch.IncomingTrades <- exchange.NewTransaction(
				exchange.BuyTransactionType,
				exchange.TransactionAmtDataType(
					getRandomIntForBuy(
						int(stkExch.LastTradedPrice),
					),
				),
			)
			stkExch.IncomingTrades <- exchange.NewTransaction(
				exchange.SellTransactionType,
				exchange.TransactionAmtDataType(
					getRandomIntForSell(
						int(stkExch.LastTradedPrice),
					),
				),
			)
		}
	}
}

func getRandomIntForBuy(target int) int {
	min := (target - 100)
	max := target
	return rand.Intn(max-min+1) + min
}

func getRandomIntForSell(target int) int {
	min := target - 25
	max := target + 100
	return rand.Intn(max-min+1) + min
}

func blockUntilSigInt() {
	// Create a channel to receive OS signals.
	signalChannel := make(chan os.Signal, 1)

	// Notify the signal channel on receiving SIGINT (Ctrl+C) or SIGTERM signals.
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Press Ctrl+C to exit...")
	// Block until a signal is received.
	<-signalChannel
	// Handle the received signal.
	fmt.Printf("Exiting...\n")
}
