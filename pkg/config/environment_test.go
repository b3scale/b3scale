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

func TestDomainOf(t *testing.T) {
	host := "https://cluster.bbb.foo.example"
	if DomainOf(host) != "foo.example" {
		t.Error("DomainOf failed:", DomainOf(host))
	}
	host = "https://cluster.bbb.foo.example/"
	if DomainOf(host) != "foo.example" {
		t.Error("DomainOf failed:", DomainOf(host))
	}
	host = "https://cluster.bbb.foo.example:8080"
	if DomainOf(host) != "foo.example" {
		t.Error("DomainOf failed:", DomainOf(host))
	}
	host = "foo.example"
	if DomainOf(host) != "foo.example" {
		t.Error("DomainOf failed:", host, DomainOf(host))
	}
	host = "foo"
	if DomainOf(host) != "foo" {
		t.Error("DomainOf failed:", host, DomainOf(host))
	}

}
