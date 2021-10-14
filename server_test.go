package fsync

import (
	"testing"
)

func TestServe(t *testing.T) {
	s := NewServer("localhost:8081")

	if s == nil {
		t.Fatal("expect not a nil")
	}
	s.Stop()
}
