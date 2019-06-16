package dkvs

import (
	"testing"
	"time"
)

// Test the transport by keeping it open for 2 seconds
func TestTransport(t *testing.T) {
	tp := NewHTTPTransport()

	go func() {
		err := tp.Start(&Node{Address: ":2345"})
		if err != nil {
			t.Errorf("failed to start transport with error: %v", err)
			return
		}
	}()

	time.Sleep(2 * time.Second)

	err := tp.Stop()
	if err != nil {
		t.Errorf("failed to stop transport with error: %v", err)
		return
	}
}
