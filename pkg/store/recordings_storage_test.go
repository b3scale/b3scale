package store

import (
	"os"
	"testing"
)

func TestRecordingStorageCheckPath(t *testing.T) {
	if err := os.Mkdir("./presentation", 755); err != nil {
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
