package aesbnc

import (
	"encoding/hex"
	"testing"
)

func TestKnownVector(t *testing.T) {
	enc, err := Encrypt([]byte("hello-cover"))
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(enc); got != "0c849758042af1c9174d3b41c36c7f02" {
		t.Fatalf("got %s", got)
	}
}

func TestToBncKey(t *testing.T) {
	if ToBncKey("a/cover.jpg") != "a/cover.bnc" {
		t.Fatal(ToBncKey("a/cover.jpg"))
	}
	if ToBncKey("a/cover.bnc") != "a/cover.bnc" {
		t.Fatal(ToBncKey("a/cover.bnc"))
	}
}
