package main

import (
	"testing"
	"time"
	
	"github.com/rohan/stock-simulator/exchange"
)

// TestMainInitialization tests that the main package initializes correctly
// This is a basic smoke test to ensure the main functions don't panic
func TestMainInitialization(t *testing.T) {
	// Test the random price generation functions
	testCases := []struct {
		name       string
		targetPrice int
		minExpected int
		maxExpected int
		testFunc    func(int) int
	}{
		{
			name:       "Buy price generation",
			targetPrice: 100,
			minExpected: 1,
			maxExpected: 100,
			testFunc:    getRandomIntForBuy,
		},
		{
			name:       "Sell price generation",
			targetPrice: 100,
			minExpected: 75,
			maxExpected: 200,
			testFunc:    getRandomIntForSell,
		},
		{
			name:       "Buy price generation with low target",
			targetPrice: 10,
			minExpected: 1,
			maxExpected: 10,
			testFunc:    getRandomIntForBuy,
		},
		{
			name:       "Sell price generation with low target",
			targetPrice: 10,
			minExpected: 1,
			maxExpected: 110,
			testFunc:    getRandomIntForSell,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the function multiple times to check range
			for i := 0; i < 100; i++ {
				result := tc.testFunc(tc.targetPrice)
				
				// Check that the result is within the expected range
				if result < tc.minExpected || result > tc.maxExpected {
					t.Errorf("Expected result between %d and %d, got %d", 
						tc.minExpected, tc.maxExpected, result)
				}
			}
		})
	}
}

// TestGenerateRandomTradesDoesNotPanic tests that the generateRandomTrades function doesn't panic
func TestGenerateRandomTradesDoesNotPanic(t *testing.T) {
	// This is a very basic test that just ensures the function doesn't panic
	// We'll create a mock exchange and logger and run the function for a short time
	
	// Create a mock exchange
	mockExchange := exchange.NewExchange(100)
	mockLogger := exchange.NewLogger("TestLogger")
	
	// Start the function in a goroutine
	done := make(chan bool)
	go func() {
		// We'll catch any panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("generateRandomTrades panicked: %v", r)
			}
			done <- true
		}()
		
		// Start the function
		go generateRandomTrades(&mockExchange, mockLogger)
		
		// Let it run for a short time
		time.Sleep(100 * time.Millisecond)
	}()
	
	// Wait for the test to complete
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(1 * time.Second):
		t.Errorf("Test timed out")
	}
}
