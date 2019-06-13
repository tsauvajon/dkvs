package main

import "testing"

// Test setting then getting a value
func TestSetGet(t *testing.T) {
	s := NewStore()

	key := "testkey123"
	value := "hello"
	expected := "hello"

	if err := s.Set(key, value); err != nil {
		t.Errorf("setting failed: %v", err)
		return
	}

	actual, err := s.Get(key)

	if err != nil {
		t.Errorf("setting failed: %v", err)
		return
	}

	if string(actual) != expected {
		t.Errorf("expected %s, got %s", expected, string(actual))
	}
}

// Test failing on inexisting keys
func TestNotFound(t *testing.T) {
	s := NewStore()

	setkey := "testkey123"
	getkey := "thisKeyDoesntExist"
	value := "hello"

	if err := s.Set(setkey, value); err != nil {
		t.Errorf("setting failed: %v", err)
		return
	}

	actual, err := s.Get(getkey)

	if err == nil {
		t.Errorf("should have failed, instead found value: %s", string(actual))
	}
}
