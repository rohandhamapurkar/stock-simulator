package ui

import (
	"encoding/json"
	"net/http"
	"stockmarketsim/exchange"
	"time"
)

// Server represents the UI server
type Server struct {
	wsManager *exchange.WebSocketManager
	exchange  *exchange.Exchange
	logger    *exchange.Logger
}

// NewServer creates a new UI server
func NewServer(exch *exchange.Exchange) *Server {
	return &Server{
		wsManager: exchange.NewWebSocketManager(),
		exchange:  exch,
		logger:    exchange.NewLogger("UIServer"),
	}
}

// Start starts the UI server
func (s *Server) Start(port string) {
	// Serve static files from the ui/static directory
	http.Handle("/", http.FileServer(http.Dir("ui/static")))

	// WebSocket endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.wsManager.HandleWebSocket(w, r, s.exchange)
	})

	// API endpoint to get the current price
	http.HandleFunc("/api/price", func(w http.ResponseWriter, r *http.Request) {
		price := s.exchange.LastTradedPrice
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"price": price,
		})
	})

	// API endpoint to get price history
	http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		history := s.wsManager.GetPriceHistory()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	})

	// API endpoint to get the current order book
	http.HandleFunc("/api/orderbook", func(w http.ResponseWriter, r *http.Request) {
		orderBook := s.exchange.GetOrderBook()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orderBook)
	})

	// Start the server
	s.logger.Info("Starting UI server on port " + port)
	go func() {
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			s.logger.Fatal("Failed to start UI server: " + err.Error())
		}
	}()

	// Start a goroutine to periodically broadcast the order book
	go s.broadcastOrderBookPeriodically()
}

// broadcastOrderBookPeriodically broadcasts the order book every second
func (s *Server) broadcastOrderBookPeriodically() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		<-ticker.C
		orderBook := s.exchange.GetOrderBook()
		s.wsManager.BroadcastOrderBook(orderBook)
	}
}

// BroadcastPriceUpdate broadcasts a price update to all connected clients
func (s *Server) BroadcastPriceUpdate(price int) {
	s.wsManager.BroadcastPriceUpdate(price)
}
