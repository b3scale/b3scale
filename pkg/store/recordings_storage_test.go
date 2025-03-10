package store

import (
	"os"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
)

func TestRecordingStorageCheckPath(t *testing.T) {
	if err := os.Mkdir("./presentation", 0755); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll("./presentation"); err != nil {
			t.Fatal(err)
		}
	}()

	s := &RecordingsStorage{
		PublishedPath:   ".",
		UnpublishedPath: ".",
	}
	if err := s.Check(); err != nil {
		t.Error(err)
	}

	s.PublishedPath = "/nfo0aiub3ui12bivv"
	if err := s.Check(); err == nil {
		t.Error("expected error")
	} else {
		t.Log(err)
	}
}

func TestRecordingStoreageListThumbnailFiles(t *testing.T) {
	s := &RecordingsStorage{
		PublishedPath:   "../../testdata/recordings/published",
		UnpublishedPath: "../../testdata/recordings/unpublished",
	}

	id := "f8bedf660bfa3604f9b6c63fe37c8a85d46e8e90-1647280741542"
	thumbnails := s.ListThumbnailFiles(id)
	t.Log(thumbnails)

	if len(thumbnails) != 6 {
		t.Error("there should be 6 thumbnails, found:", len(thumbnails), thumbnails)
	}
}

func TestRecordingsStorageMakeRecordingPreview(t *testing.T) {
	s := &RecordingsStorage{
		PublishedPath:   "../../testdata/recordings/published",
		UnpublishedPath: "../../testdata/recordings/unpublished",
	}
	rec := &bbb.Recording{
		RecordID: "f8bedf660bfa3604f9b6c63fe37c8a85d46e8e90-1647280741542",
	}
	preview := s.MakeRecordingPreview(rec)

	previewURL := preview.Images.All[0].URL
	expected := "presentation/f8bedf660bfa3604f9b6c63fe37c8a85d46e8e90-1647280741542/presentation/d7c0bd2c86d5fcf83cdc78dcfbdfc0c83495d17e-1647280741542/thumbnails/thumb-1.png"

	if previewURL != expected {
		t.Error("unexpected previewURL", previewURL, "expected:", expected)
	}
}

// Test Filesystem handling
func TestAssertFsSafe(t *testing.T) {
	// Valid
	v := "f0fa3bb0d10478f00-bar-21238912839"
	if err := assertFsSafe(v); err != nil {
		t.Error(err)
	}

	v = "presentation"
	if err := assertFsSafe(v); err != nil {
		t.Error(err)
	}

	// Invalid
	v = ""
	if err := assertFsSafe(v); err == nil {
		t.Error("may not be empty")
	}

	v = ".."
	if err := assertFsSafe(v); err == nil {
		t.Error("expected error for ..")
	}

	v = "foo/bar"
	if err := assertFsSafe(v); err == nil {
		t.Error("expected error for /")
	}

	v = "foo\\bar"
	if err := assertFsSafe(v); err == nil {
		t.Error("may not contain \\")
	}

	v = "foo\nbar"
	if err := assertFsSafe(v); err == nil {
		t.Error("may not contain newline")
	}

	v = "foo\rbar"
	if err := assertFsSafe(v); err == nil {
		t.Error("may not contain control chars")
	}
}
