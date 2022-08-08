package config

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
	max := len(list) - 2
	ref := make([]string, n)
	for i := 0; i < n; i++ {
		ref[i] = list[rand.Intn(max)]
	}
	return strings.Join(ref, "-")
}
