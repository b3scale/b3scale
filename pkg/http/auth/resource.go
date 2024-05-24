package auth

import (
	"fmt"
	"strings"
)

func EncodeResource(hint, id string) string {
	return fmt.Sprintf("%s:%s", hint, id)
}

func MustDecodeResource(resource string) (string, string) {
	t := strings.SplitN(resource, ":", 2)
	if len(t) < 2 {
		panic("invalid resource token")
	}
	return t[1], t[0]
}
