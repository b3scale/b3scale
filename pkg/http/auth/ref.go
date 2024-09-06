package auth

import (
	_ "embed" // required for embedding
	"math/rand"

	"strings"
)

//go:embed words.txt
var words string

// GenerateRef will create a most likely unique
// combination of words.
func GenerateRef(n int) string {
	list := strings.Split(words, "\n")
	max := len(list) - 1 // last is ""
	ref := make([]string, n)
	for i := 0; i < n; i++ {
		ref[i] = list[rand.Intn(max)] // i <= n < max
	}
	return strings.Join(ref, "-")
}

var sigma = []rune("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ-_")

// GenerateNonce will create a random string of length n
func GenerateNonce(n int) string {
	nonce := make([]rune, n)
	max := len(sigma)
	for i := 0; i < n; i++ {
		nonce[i] = sigma[rand.Intn(max)]
	}
	return string(nonce)
}
