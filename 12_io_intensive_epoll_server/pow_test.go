package main

import (
	"testing"
	"time"
)

func TestPow(t *testing.T) {
	for i := 6; i < 8; i++ {
		now := time.Now()
		hash := pow(i)
		t.Logf("target: %d, took %d ms, hash: %v\n", i, time.Since(now), hash)
	}
}
