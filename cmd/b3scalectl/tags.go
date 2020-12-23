package main

import (
	"sort"
	"strings"
)

// Check if list of tags is equal
func tagsEq(t1, t2 []string) bool {
	sort.Strings(t1)
	sort.Strings(t2)
	return strings.Join(t1, "") == strings.Join(t2, "")
}
