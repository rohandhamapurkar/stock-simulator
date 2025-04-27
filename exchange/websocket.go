package exchange

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType defines the type of message sent over WebSocket
type MessageType string

const (
	// PriceUpdateMessage is sent when the price changes
	PriceUpdateMessage MessageType = "price_update"
	// OrderBookMessage is sent when the order book changes
	OrderBookMessage MessageType = "order_book"
)

// WebSocketMessage is the base structure for all messages sent over WebSocket
type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// PriceUpdate represents a price update message sent to clients
type PriceUpdate struct {
	Price int `json:"price"`
}

// WebSocketManager manages WebSocket connections and broadcasts updates
type WebSocketManager struct {
	clients      map[*websocket.Conn]bool
	clientsMutex sync.Mutex
	upgrader     websocket.Upgrader
	priceHistory []WebSocketMessage
	historyMutex sync.Mutex
	logger       *Logger
}

// NewWebSocketManager creates a new WebSocketManager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:      make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow connections from any origin for development
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		priceHistory: make([]WebSocketMessage, 0, 100),
		logger:       NewLogger("WebSocket"),
	}
}

// HandleWebSocket handles WebSocket connections
func (wsm *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request, exchange *Exchange) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := wsm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsm.logger.Error("Failed to upgrade connection: " + err.Error())
		return
	}

	// Register the new client
	wsm.clientsMutex.Lock()
	wsm.clients[conn] = true
	wsm.clientsMutex.Unlock()

	wsm.logger.Info("New client connected")

	// Send the price history to the new client
	wsm.historyMutex.Lock()
	if len(wsm.priceHistory) > 0 {
		historyJSON, err := json.Marshal(wsm.priceHistory)
		if err == nil {
			conn.WriteMessage(websocket.TextMessage, historyJSON)
		}
	}
	wsm.historyMutex.Unlock()

	// Send the current order book to the new client
	if exchange != nil {
		orderBook := exchange.GetOrderBook()
		message := WebSocketMessage{
			Type:      OrderBookMessage,
			Timestamp: time.Now(),
			Data:      orderBook,
		}

		messageJSON, err := json.Marshal(message)
		if err == nil {
			conn.WriteMessage(websocket.TextMessage, messageJSON)
		}
	}

	// Handle disconnections
	go func() {
		for {
			// Read messages from the client (we don't actually use them, but need to detect disconnections)
			_, _, err := conn.ReadMessage()
			if err != nil {
				wsm.clientsMutex.Lock()
				delete(wsm.clients, conn)
				wsm.clientsMutex.Unlock()
				conn.Close()
				wsm.logger.Info("Client disconnected")
				break
			}
		}
	}()
}

// BroadcastPriceUpdate broadcasts a price update to all connected clients
func (wsm *WebSocketManager) BroadcastPriceUpdate(price int) {
	priceData := PriceUpdate{
		Price: price,
	}

	message := WebSocketMessage{
		Type:      PriceUpdateMessage,
		Timestamp: time.Now(),
		Data:      priceData,
	}

	// Add to price history
	wsm.historyMutex.Lock()
	wsm.priceHistory = append(wsm.priceHistory, message)
	// Keep only the last 100 price updates
	if len(wsm.priceHistory) > 100 {
		wsm.priceHistory = wsm.priceHistory[len(wsm.priceHistory)-100:]
	}
	wsm.historyMutex.Unlock()

	// Marshal the message to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		wsm.logger.Error("Failed to marshal price update: " + err.Error())
		return
	}

	// Broadcast to all clients
	wsm.clientsMutex.Lock()
	for client := range wsm.clients {
		err := client.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			wsm.logger.Warn("Error sending to client: " + err.Error())
			client.Close()
			delete(wsm.clients, client)
		}
	}
	wsm.clientsMutex.Unlock()
}

// GetPriceHistory returns the price history
func (wsm *WebSocketManager) GetPriceHistory() []WebSocketMessage {
	wsm.historyMutex.Lock()
	defer wsm.historyMutex.Unlock()

	// Return a copy to avoid race conditions
	history := make([]WebSocketMessage, len(wsm.priceHistory))
	copy(history, wsm.priceHistory)
	return history
}

// BroadcastOrderBook broadcasts the current order book to all connected clients
func (wsm *WebSocketManager) BroadcastOrderBook(orderBook OrderBook) {
	message := WebSocketMessage{
		Type:      OrderBookMessage,
		Timestamp: time.Now(),
		Data:      orderBook,
	}

	// Marshal the message to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		wsm.logger.Error("Failed to marshal order book: " + err.Error())
		return
	}

	// Broadcast to all clients
	wsm.clientsMutex.Lock()
	for client := range wsm.clients {
		err := client.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			wsm.logger.Warn("Error sending to client: " + err.Error())
			client.Close()
			delete(wsm.clients, client)
		}
	}
	wsm.clientsMutex.Unlock()
}
