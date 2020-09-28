package cluster

import (
	"testing"
)

func TestSelectFirst(t *testing.T) {
	b := []*Backend{
		&Backend{},
		&Backend{},
		&Backend{},
	}
	if len(selectFirst(b)) != 1 {
		t.Error("Expected only one backend.")
	}

	b = []*Backend{}
	if len(selectFirst(b)) != 0 {
		t.Error("No backends should not be a problem.")
	}
}
