package store

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

var (
	// ErrRecordingsStorageUnconfigured will be returned, when
	// the environment variables for the published and unpublished
	// recording path are missing.
	ErrRecordingsStorageUnconfigured = errors.New(
		"environment for " + config.EnvRecordingsPublishedPath + " or " +
			config.EnvRecordingsUnpublishedPath + " is not set")
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
	publishedPath, ok := config.GetEnvOpt(config.EnvRecordingsPublishedPath)
	if !ok {
		return nil, ErrRecordingsStorageUnconfigured
	}
	unpublishedPath, ok := config.GetEnvOpt(config.EnvRecordingsUnpublishedPath)
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

	// Strip base path
	thumbnails := make([]string, 0, len(th))
	prefix := s.PublishedRecordingPath(recordID)
	for _, t := range th {
		thumbnails = append(thumbnails, t[len(prefix)+1:])
	}

	return thumbnails
}

// MakeRecordingPreview will use the thumbnails to create previews
func (s *RecordingsStorage) MakeRecordingPreview(
	recordID string,
) *bbb.Preview {
	thumbnails := s.ListThumbnailFiles(recordID)
	images := make([]*bbb.Image, 0, len(thumbnails))

	for i, th := range thumbnails {
		img := &bbb.Image{
			URL: path.Join("presentation", recordID, th),
			Alt: fmt.Sprintf("Thumbnail %02d", i+1),
		}
		images = append(images, img)
	}

	p := &bbb.Preview{
		Images: &bbb.Images{
			All: images,
		},
	}
	return p
}
