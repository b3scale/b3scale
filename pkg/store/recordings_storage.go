package store

import (
	"errors"
	"os"
	"path/filepath"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

var (
	// ErrRecordingsStorageUnconfigured will be returned, when
	// the environment variables for the published and unpublished
	// recording path are missing.
	ErrRecordingsStorageUnconfigured = errors.New(
		"environment for " + config.EnvPublishedRecordingsPath + " or " +
			config.EnvUnpublishedRecordingsPath + " is not set")
)

// RecordingsStorage is handleing the filesystem access
// when dealing with recordings.
type RecordingsStorage struct {
	PublishedPath   string
	UnpublishedPath string
}

// NewRecordingsStorageFromEnv creates a new recordings storage
// instance and configures it through well known environment variables.
func NewRecordingsStorageFromEnv() (*RecordingsStorage, error) {
	publishedPath, ok := config.GetEnvOpt(config.EnvPublishedRecordingsPath)
	if !ok {
		return nil, ErrRecordingsStorageUnconfigured
	}
	unpublishedPath, ok := config.GetEnvOpt(config.EnvUnpublishedRecordingsPath)
	if !ok {
		return nil, ErrRecordingsStorageUnconfigured
	}
	s := &RecordingsStorage{
		PublishedPath:   publishedPath,
		UnpublishedPath: unpublishedPath,
	}
	return s, nil
}

// PublishedRecordingPath returns the joined filepath
// for an "id" (this will be the internal meeting id).
func (s *RecordingsStorage) PublishedRecordingPath(id string) string {
	return filepath.Join(s.PublishedPath, "presentation", id)
}

// UnpublishedRecordingPath returns the joined filepath
// for an "id" (this will be the internal meeting id).
func (s *RecordingsStorage) UnpublishedRecordingPath(id string) string {
	return filepath.Join(s.UnpublishedPath, "presentation", id)
}

// Check will test if we can access and manipulate the
// recordings storage.
func (s *RecordingsStorage) Check() error {
	p := s.PublishedRecordingPath(".rwtest.b3scale")
	if err := checkPath(p); err != nil {
		return err
	}
	p = s.UnpublishedRecordingPath(".rwtest.b3scale")
	if err := checkPath(p); err != nil {
		return err
	}
	return nil
}

// private checkPath will test if the path is
// read and writable.
func checkPath(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil // yay
}

// ListThumbnailFiles retrievs all thumbnail files from presentations
// relative to the published path.
func (s *RecordingsStorage) ListThumbnailFiles(recordID string) []string {
	th, _ := filepath.Glob(
		filepath.Join(
			s.PublishedRecordingPath(recordID),
			"presentation", "*", "thumbnails", "*.png"))

	// Strip recording path
	thumbnails := make([]string, 0, len(th))
	prefix := s.PublishedRecordingPath(recordID)
	for _, t := range th {
		thumbnails = append(thumbnails, t[len(prefix)+1:])
	}

	return thumbnails
}
