package main

import "testing"

func TestFail(t *testing.T) {
	t.Error("intentionally to fail the test")
}
