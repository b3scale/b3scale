package auth

import (
	"fmt"
	"strings"
)

// EncodeResource associates an ID string with a
// type hint. Resources can for example be `recordings`,
// `frontend`, etc...
func EncodeResource(hint, id string) string {
	return fmt.Sprintf("%s:%s", hint, id)
}

// MustDecodeResource decodes splits a resource token
// into a type hint and an ID string. If the token is
// invalid, this function will panic.
func MustDecodeResource(resource string) (string, string) {
	t := strings.SplitN(resource, ":", 2)
	if len(t) < 2 {
		panic("invalid resource token")
	}
	return t[1], t[0]
}
