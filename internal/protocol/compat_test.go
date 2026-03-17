package protocol

import "testing"

func TestNegotiateVersion(t *testing.T) {
	v, err := NegotiateVersion(1, 3, 1, 4)
	if err != nil {
		t.Fatalf("expected success, got err=%v", err)
	}
	if v != 3 {
		t.Fatalf("expected version 3, got %d", v)
	}
}

func TestNegotiateVersionNoOverlap(t *testing.T) {
	_, err := NegotiateVersion(4, 5, 1, 3)
	if err == nil {
		t.Fatal("expected negotiation error")
	}
}
