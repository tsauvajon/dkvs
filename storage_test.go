package dkvs

import "testing"

// Test setting then getting a value
func TestSetGet(t *testing.T) {
	s := NewStore()

	type testCase struct {
		key, value, expected string
	}

	testCases := []*testCase{
		&testCase{
			key:      "testkey123",
			value:    "hello",
			expected: "hello",
		},
		&testCase{
			key:      "q",
			value:    "hello",
			expected: "hello",
		},
		&testCase{
			key:      "testkey123",
			value:    "hello1",
			expected: "hello1",
		},
		&testCase{
			key:      "testkey123",
			value:    "asdsad",
			expected: "asdsad",
		},
	}

	for _, test := range testCases {
		if err := s.Set(test.key, test.value); err != nil {
			t.Errorf("setting failed: %v", err)
			return
		}

		actual, err := s.Get(test.key)

		if err != nil {
			t.Errorf("setting failed: %v", err)
			return
		}

		if string(actual) != test.expected {
			t.Errorf("expected %s, got %s", test.expected, string(actual))
		}
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
