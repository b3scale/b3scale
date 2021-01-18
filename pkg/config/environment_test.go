package config

import (
	"testing"
)

// TestIsEnabled tests IsEnabled
func TestIsEnabled(t *testing.T) {
	if IsEnabled("1") == false {
		t.Error("1 should be true")
	}
	if IsEnabled("yes") == false {
		t.Error("yes should be true")
	}
	if IsEnabled("true") == false {
		t.Error("true should be true")
	}
	if IsEnabled("no") == true {
		t.Error("no should be false")
	}
}
