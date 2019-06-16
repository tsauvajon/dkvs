package dkvs

import (
	"testing"
	"time"
)

// Test instantiating masters and slaves
func TestNewNode(t *testing.T) {
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

	time.Sleep(100 * time.Millisecond)

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

	time.Sleep(100 * time.Millisecond)
}

func TestIsolatedNode(t *testing.T) {
	_, err := NewSlave(":9999", ":123")
	if err == nil {
		t.Error("slaves should not be created without a valid master")
	}
}
