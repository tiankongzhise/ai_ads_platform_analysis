package store

import "testing"

func TestEncodeRESP(t *testing.T) {
	got := encodeRESP("GET", "state")
	want := "*2\r\n$3\r\nGET\r\n$5\r\nstate\r\n"
	if got != want {
		t.Fatalf("unexpected RESP encoding:\nwant %q\ngot  %q", want, got)
	}
}
