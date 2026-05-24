package store

import "testing"

func TestEncodeRESP(t *testing.T) {
	got := encodeRESP("SET", "key", "value", "EX", "60")
	want := "*5\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n$2\r\nEX\r\n$2\r\n60\r\n"
	if got != want {
		t.Fatalf("unexpected RESP encoding:\nwant %q\ngot  %q", want, got)
	}
}
