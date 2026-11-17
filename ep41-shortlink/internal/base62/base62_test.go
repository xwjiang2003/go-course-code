package base62

import "testing"

func TestEncodeDecode(t *testing.T) {
	cases := []int64{0, 1, 61, 62, 1000, 1000000, 9876543210}
	for _, n := range cases {
		s := Encode(n)
		if got := Decode(s); got != n {
			t.Errorf("Decode(Encode(%d)) = %d, want %d", n, got, n)
		}
	}
}

func TestEncodeKnown(t *testing.T) {
	if Encode(0) != "0" {
		t.Errorf("Encode(0) = %q, want 0", Encode(0))
	}
	if Encode(62) != "10" {
		t.Errorf("Encode(62) = %q, want 10", Encode(62))
	}
}
