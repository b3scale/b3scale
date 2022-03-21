package store

import (
	"errors"
	"os"

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

// NewRecordingStorageFromEnv creates a new recordings storage
// instance and configures it through well known environment variables.
func NewRecordingStorageFromEnv() (*RecordingsStorage, error) {
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

// Check will test if we can access and manipulate the
// recordings storage.
func (s *RecordingsStorage) Check() error {
	// Test: Can read
	if err := checkPath(s.PublishedPath); err != nil {
		return err
	}
	if err := checkPath(s.UnpublishedPath); err != nil {
		return err
	}
	return nil
}

// private checkPath will test if the path is
// read and writable.
func checkPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
