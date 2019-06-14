package dkvs

import "testing"

// Test instantiating masters and slaves
func TestNewNode(t *testing.T) {
	m, err := NewMaster()

	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	if m.id == "" {
		t.Error("created nodes should have an id")
		return
	}

	s, err := NewSlave("")

	if err != nil {
		t.Errorf("creating a slave failed with error: %v", err)
		return
	}

	if s.id == "" {
		t.Error("created nodes should have an id")
		return
	}
}
