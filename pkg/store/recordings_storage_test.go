package store

import "testing"

func TestRecordingStorageCheckPath(t *testing.T) {
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
