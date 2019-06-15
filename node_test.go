package dkvs

import (
	"os"
	"testing"
	"time"
)

// Test instantiating masters and slaves
func TestNewNode(t *testing.T) {
	// do not print "server closed" messages on the console
	os.Stdout, _ = os.Open(os.DevNull)
	m, err := NewMaster(":1234")
	if m != nil {
		defer m.Close()
	}

	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	if m.ID == "" {
		t.Error("created nodes should have an id")
		return
	}

	time.Sleep(1 * time.Second)

	s, err := NewSlave(":1235", ":1234")
	if s != nil {
		defer s.Close()
	}

	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	if s.ID == "" {
		t.Error("created nodes should have an id")
		return
	}

	time.Sleep(1 * time.Second)
}
