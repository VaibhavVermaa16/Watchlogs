package main

import "testing"

func TestHello(t *testing.T) {
	expected := "Hello, World!"
	if result := Hello(); result != expected {
		t.Errorf("Hello() = %q; want %q", result, expected)
	}
}
