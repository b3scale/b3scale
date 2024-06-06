package auth

import "testing"

func TestGenerateRef(t *testing.T) {
	for i := 0; i < 15; i++ {
		t.Log(GenerateRef(4))
	}
}

func TestGenerateNonce(t *testing.T) {
	for i := 0; i < 15; i++ {
		t.Log(GenerateNonce(24))
	}
}
