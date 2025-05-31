package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rohan/stock-simulator/exchange"
)

func TestNewServer(t *testing.T) {
	// Create a test exchange
	exch := exchange.NewExchange(100)

	// Create a new server
	server := NewServer(&exch)

	// Check that the server is properly initialized
	if server.exchange == nil {
		t.Errorf("Expected exchange to be initialized")
	}

	if server.wsManager == nil {
		t.Errorf("Expected WebSocket manager to be initialized")
	}

	if server.logger == nil {
		t.Errorf("Expected logger to be initialized")
	}
}

func TestAPIEndpoints(t *testing.T) {
	// Create a test exchange
	exch := exchange.NewExchange(100)

	// Add some orders to the exchange
	buyTxn := exchange.NewTransaction(exchange.BuyTransactionType, 90)
	sellTxn := exchange.NewTransaction(exchange.SellTransactionType, 110)
	exch.BuyQ.Insert(buyTxn)
	exch.SellQ.Insert(sellTxn)

	// Create a new server
	server := NewServer(&exch)

	// Test cases for API endpoints
	testCases := []struct {
		name           string
		endpoint       string
		expectedStatus int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name:           "Price API",
			endpoint:       "/api/price",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				price, ok := response["price"]
				if !ok {
					t.Errorf("Expected 'price' field in response")
					return
				}

				// The price should be a number (float64 in JSON)
				_, ok = price.(float64)
				if !ok {
					t.Errorf("Expected price to be a number, got %T", price)
				}
			},
		},
		{
			name:           "History API",
			endpoint:       "/api/history",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response []interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				// The response should be an array (possibly empty)
				if response == nil {
					t.Errorf("Expected array response, got nil")
				}
			},
		},
		{
			name:           "Order Book API",
			endpoint:       "/api/orderbook",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				// Check for expected fields
				buyOrders, ok := response["buyOrders"]
				if !ok {
					t.Errorf("Expected 'buyOrders' field in response")
				}

				sellOrders, ok := response["sellOrders"]
				if !ok {
					t.Errorf("Expected 'sellOrders' field in response")
				}

				// Check that the orders are arrays
				_, ok = buyOrders.([]interface{})
				if !ok {
					t.Errorf("Expected buyOrders to be an array, got %T", buyOrders)
				}

				_, ok = sellOrders.([]interface{})
				if !ok {
					t.Errorf("Expected sellOrders to be an array, got %T", sellOrders)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request to the endpoint
			req, err := http.NewRequest("GET", tc.endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Create a handler for the endpoint
			var handler http.HandlerFunc
			switch tc.endpoint {
			case "/api/price":
				handler = func(w http.ResponseWriter, r *http.Request) {
					price := server.exchange.LastTradedPrice
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"price": price,
					})
				}
			case "/api/history":
				handler = func(w http.ResponseWriter, r *http.Request) {
					history := server.wsManager.GetPriceHistory()
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(history)
				}
			case "/api/orderbook":
				handler = func(w http.ResponseWriter, r *http.Request) {
					orderBook := server.exchange.GetOrderBook()
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(orderBook)
				}
			}

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, status)
			}

			// Validate the response
			tc.validateResponse(t, rr.Body.Bytes())
		})
	}
}

func TestBroadcastPriceUpdate(t *testing.T) {
	// Create a test exchange
	exch := exchange.NewExchange(100)

	// Create a new server
	server := NewServer(&exch)

	// Test broadcasting a price update
	testPrice := 150
	
	// This should not panic
	server.BroadcastPriceUpdate(testPrice)
}

func TestBroadcastOrderBookPeriodically(t *testing.T) {
	// This test is more of a smoke test to ensure the function doesn't panic
	// Create a test exchange
	exch := exchange.NewExchange(100)

	// Create a new server
	server := NewServer(&exch)

	// Start the broadcast goroutine
	go server.broadcastOrderBookPeriodically()

	// Wait a short time to allow at least one broadcast
	time.Sleep(100 * time.Millisecond)

	// No assertions needed - we're just checking that it doesn't panic
}
