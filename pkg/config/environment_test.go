package config

import (
	"os"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
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

func TestGetRecordingsDefaultVisibility(t *testing.T) {
	// Default
	os.Unsetenv(EnvRecordingsDefaultVisibility)
	v := GetRecordingsDefaultVisibility()
	if v != bbb.RecordingVisibilityPublished {
		t.Error("unexpected:", v)
	}

	// From env
	os.Setenv(EnvRecordingsDefaultVisibility, "public_protected")
	v = GetRecordingsDefaultVisibility()
	if v != bbb.RecordingVisibilityPublicProtected {
		t.Error("unexpected:", v)
	}
}

func TestGetRecordingsInboxPath(t *testing.T) {
	// Prepare env
	os.Unsetenv(EnvRecordingsDefaultVisibility)
	os.Setenv(EnvRecordingsPublishedPath, "published")
	os.Setenv(EnvRecordingsUnpublishedPath, "unpublished")
	os.Setenv(EnvRecordingsInboxPath, "inbox")

	p := GetRecordingsInboxPath()
	if p != "inbox" {
		t.Error("unexpected inbox path:", p)
	}

	os.Unsetenv(EnvRecordingsInboxPath)
	p = GetRecordingsInboxPath()
	if p != "published" {
		t.Error("expected published, got", p)
	}

	os.Setenv(EnvRecordingsDefaultVisibility, "unpublished")
	p = GetRecordingsInboxPath()
	if p != "unpublished" {
		t.Error("expected unpublished, got", p)
	}
}
