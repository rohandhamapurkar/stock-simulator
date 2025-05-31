package exchange

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewWebSocketManager(t *testing.T) {
	wsm := NewWebSocketManager()

	// Check that the manager is properly initialized
	if wsm.clients == nil {
		t.Errorf("Expected clients map to be initialized")
	}

	if wsm.priceHistory == nil {
		t.Errorf("Expected priceHistory slice to be initialized")
	}

	if wsm.logger == nil {
		t.Errorf("Expected logger to be initialized")
	}
}

func TestBroadcastPriceUpdate(t *testing.T) {
	wsm := NewWebSocketManager()

	// Test broadcasting a price update
	testPrice := 150
	wsm.BroadcastPriceUpdate(testPrice)

	// Check that the price history was updated
	if len(wsm.priceHistory) != 1 {
		t.Errorf("Expected 1 entry in price history, got %d", len(wsm.priceHistory))
	}

	if len(wsm.priceHistory) > 0 {
		message := wsm.priceHistory[0]
		if message.Type != PriceUpdateMessage {
			t.Errorf("Expected message type %s, got %s", PriceUpdateMessage, message.Type)
		}

		// Check the data
		priceData, ok := message.Data.(PriceUpdate)
		if !ok {
			t.Errorf("Expected data to be of type PriceUpdate")
		} else if priceData.Price != testPrice {
			t.Errorf("Expected price %d, got %d", testPrice, priceData.Price)
		}
	}
}

func TestGetPriceHistory(t *testing.T) {
	wsm := NewWebSocketManager()

	// Add some price updates
	prices := []int{100, 105, 110, 115, 120}
	for _, price := range prices {
		wsm.BroadcastPriceUpdate(price)
	}

	// Get the price history
	history := wsm.GetPriceHistory()

	// Check that we got the expected number of entries
	if len(history) != len(prices) {
		t.Errorf("Expected %d entries in price history, got %d", len(prices), len(history))
	}

	// Check that the entries are in the correct order
	for i, message := range history {
		priceData, ok := message.Data.(PriceUpdate)
		if !ok {
			t.Errorf("Expected data to be of type PriceUpdate")
			continue
		}

		if priceData.Price != prices[i] {
			t.Errorf("Expected price %d at index %d, got %d", prices[i], i, priceData.Price)
		}
	}
}

func TestPriceHistoryLimit(t *testing.T) {
	wsm := NewWebSocketManager()

	// Add more than 100 price updates (the limit)
	for i := 0; i < 110; i++ {
		wsm.BroadcastPriceUpdate(i)
	}

	// Get the price history
	history := wsm.GetPriceHistory()

	// Check that we only have 100 entries (the limit)
	if len(history) != 100 {
		t.Errorf("Expected 100 entries in price history (the limit), got %d", len(history))
	}

	// Check that we have the most recent 100 entries (10-109)
	for i, message := range history {
		priceData, ok := message.Data.(PriceUpdate)
		if !ok {
			t.Errorf("Expected data to be of type PriceUpdate")
			continue
		}

		expectedPrice := i + 10 // We should have entries 10-109
		if priceData.Price != expectedPrice {
			t.Errorf("Expected price %d at index %d, got %d", expectedPrice, i, priceData.Price)
		}
	}
}

func TestBroadcastOrderBook(t *testing.T) {
	wsm := NewWebSocketManager()

	// Create a test order book
	orderBook := OrderBook{
		BuyOrders: []OrderBookEntry{
			{ID: "buy1", Price: 90, Type: BuyTransactionType},
			{ID: "buy2", Price: 95, Type: BuyTransactionType},
		},
		SellOrders: []OrderBookEntry{
			{ID: "sell1", Price: 105, Type: SellTransactionType},
			{ID: "sell2", Price: 110, Type: SellTransactionType},
		},
		Timestamp: time.Now(),
	}

	// Broadcast the order book
	wsm.BroadcastOrderBook(orderBook)

	// Since there are no connected clients, we can't directly test the broadcast
	// But we can verify that the method doesn't panic
}

// This is a more complex test that requires a WebSocket server and client
// It's included for completeness but may be skipped in some environments
func TestHandleWebSocket(t *testing.T) {
	// Skip this test in automated environments
	if testing.Short() {
		t.Skip("Skipping WebSocket connection test in short mode")
	}

	// Create a test exchange
	exchange := NewExchange(100)

	// Create a WebSocket manager
	wsm := NewWebSocketManager()

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsm.HandleWebSocket(w, r, &exchange)
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/"

	// Connect a test client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	// Broadcast a price update
	testPrice := 150
	wsm.BroadcastPriceUpdate(testPrice)

	// Wait for the message
	_, message, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	// Parse the message
	var wsMessage WebSocketMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		t.Fatalf("Failed to parse message: %v", err)
	}

	// Check the message type
	if wsMessage.Type != PriceUpdateMessage {
		t.Errorf("Expected message type %s, got %s", PriceUpdateMessage, wsMessage.Type)
	}

	// Check the price data
	priceData, ok := wsMessage.Data.(map[string]interface{})
	if !ok {
		t.Errorf("Expected data to be a map")
	} else {
		price, ok := priceData["price"].(float64)
		if !ok {
			t.Errorf("Expected price to be a number")
		} else if int(price) != testPrice {
			t.Errorf("Expected price %d, got %d", testPrice, int(price))
		}
	}
}
