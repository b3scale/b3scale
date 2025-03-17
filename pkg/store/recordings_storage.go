package store

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/rs/zerolog/log"
)

var (
	// ErrRecordingsStorageUnconfigured will be returned, when
	// the environment variables for the published and unpublished
	// recording path are missing.
	ErrRecordingsStorageUnconfigured = errors.New(
		"environment for " + config.EnvRecordingsPublishedPath + " or " +
			config.EnvRecordingsUnpublishedPath + " is not set")
)

// Validation Regex
var (
	ReMatchCharset = regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
)

// RecordingsStorage is handleing the filesystem access
// when dealing with recordings.
type RecordingsStorage struct {
	InboxPath       string
	PublishedPath   string
	UnpublishedPath string
}

// NewRecordingsStorageFromEnv creates a new recordings storage
// instance and configures it through well known environment variables.
func NewRecordingsStorageFromEnv() (*RecordingsStorage, error) {
	inPath := config.GetRecordingsInboxPath()
	pubPath := config.GetRecordingsPublishedPath()
	unpubPath := config.GetRecordingsUnpublishedPath()

	if pubPath == "" || unpubPath == "" {
		return nil, ErrRecordingsStorageUnconfigured
	}
	s := &RecordingsStorage{
		InboxPath:       inPath,
		PublishedPath:   pubPath,
		UnpublishedPath: unpubPath,
	}
	return s, nil
}

// Internal: assertPathAccess will test if the path is
// read- and writable.
func assertPathAccess(path string) error {
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

// Check will test if we can access and manipulate the
// recordings storage.
func (s *RecordingsStorage) Check() error {
	canary := ".rwtest.b3scale"

	// Inbox
	p := filepath.Join(s.InboxPath, canary)
	if err := assertPathAccess(p); err != nil {
		return err
	}
	// Published
	p = filepath.Join(s.PublishedPath, canary)
	if err := assertPathAccess(p); err != nil {
		return err
	}
	// Unpublished
	p = filepath.Join(s.UnpublishedPath, canary)
	if err := assertPathAccess(p); err != nil {
		return err
	}
	return nil
}

// ListThumbnailFiles retrievs all thumbnail files from presentations
// relative to the published path.
func (s *RecordingsStorage) ListThumbnailFiles(rec *bbb.Recording) []string {
	recID := rec.RecordID
	storeBasePaths := []string{
		s.InboxPath,
		s.PublishedPath,
		s.UnpublishedPath,
	}

	for _, basePath := range storeBasePaths {

		// Retrieve the thumbnail files from the presentation.
		// The recording might be located in the inbox, published or
		// unpublished path.
		formatBasePath := filepath.Join(
			basePath,
			bbb.RecordingFormatPresentation,
			recID,
		)

		th, _ := filepath.Glob(filepath.Join(
			formatBasePath, "presentation", "*", "thumbnails", "*.png"))

		// Strip base path
		thumbnails := make([]string, 0, len(th))
		for _, t := range th {
			thumbnails = append(thumbnails, t[len(formatBasePath)+1:])
		}

		if len(thumbnails) > 0 {
			return thumbnails
		}
	}

	return []string{}
}

// MakeRecordingPreview will use the thumbnails to create previews
func (s *RecordingsStorage) MakeRecordingPreview(
	rec *bbb.Recording,
) *bbb.Preview {
	recordID := rec.RecordID
	thumbnails := s.ListThumbnailFiles(rec)
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

// Filehandling
// ------------
//
// Manipulating the filesystem can be a dangerous endevour.
// Please be careful. Some tips:
//  - Validate input.
//  - Be restrictive.

// Internal: isFsSafe will check if the value contains dots
// or slashes or is non ascii. (We only expect UUID like
// hashes and timestamps.)
func assertFsSafe(v string) error {
	if v == "" {
		return fmt.Errorf("fs path component may not be empty")
	}

	// Check for non allowed chars
	if ReMatchCharset.MatchString(v) {
		return fmt.Errorf("fs patch component contains invalid chars")
	}

	return nil
}

// Internal: unsafeAssertFsPath will check if a (base) path
// exists and will try to create it.
//
// Caveat: Make sure the path is not constructed from
// _unchecked_ user input.
func unsafeAssertFsPath(p string) error {
	// Check the path exists
	if _, err := os.Stat(p); err == nil {
		return nil // nothing to do here
	}

	// Create directory
	return os.MkdirAll(p, 0755)
}

// Internal: safeDeleteRecording will delete files
// for a recording with a given format from all known paths
// Format and recordID will be validated.
func (s *RecordingsStorage) safeDeleteRecording(recordID, format string) error {
	if err := assertFsSafe(recordID); err != nil {
		return err
	}
	if err := assertFsSafe(format); err != nil {
		return err
	}

	// Published
	path := filepath.Join(s.PublishedPath, recordID, format)
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	// Unpublished
	path = filepath.Join(s.UnpublishedPath, recordID, format)
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	// Inbox
	path = filepath.Join(s.InboxPath, recordID, format)
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

// DeleteRecording will remove all files from the recording,
// for all formats of the recording.
//
// Well known formats will be checked, in case the recordings
// metadata is incomplete.
func (s *RecordingsStorage) DeleteRecording(
	rec *RecordingState,
) error {
	recID := rec.RecordID

	// Make sure to remove these well known formats
	formats := []string{
		bbb.RecordingFormatPresentation,
		bbb.RecordingFormatVideo,
		bbb.RecordingFormatPodcast,
	}

	// Add all formats from recording. As the removal of
	// a non existing directory does not result in an
	// error, we do not need to deduplicate here.
	for _, f := range rec.Recording.Formats {
		formats = append(formats, f.Type)
	}

	// Remove recording files
	for _, format := range formats {
		if err := s.safeDeleteRecording(recID, format); err != nil {
			return err
		}
		log.Debug().
			Str("recID", recID).Str("format", format).
			Msg("deleted recording files")
	}

	return nil
}

// Internal: move recording files only if dst path
// does not exist and the src _does_ exist.
//
// Note: The source might not exist, because we try
//
//	to import all known formats. Sometimes these
//	formats might not yet be ready.
//
// Caveat: Make sure src and dst are
// not constructed from _unchecked_ user input.
func unsafeMoveFiles(src, dst string) error {
	// Check the source exists
	if _, err := os.Stat(src); err != nil {
		return nil // nothing to do here
	}

	// Check if we are already present
	if _, err := os.Stat(dst); err == nil {
		log.Warn().Str("src", src).Str("dst", dst).
			Msg("not moving recording files: destination already exists")
		return nil // nothing to do here
	}

	// Move files
	return os.Rename(src, dst)
}

// Internal: saveMoveRecording will move recording
// files within the storage. Src and dst are the base paths
// within the storage. For example published, unpublished.
func (s *RecordingsStorage) saveMoveRecording(
	rec *RecordingState,
	src, dst string,
) error {
	recID := rec.RecordID
	if err := assertFsSafe(recID); err != nil {
		return err
	}

	// Move recording to published path, for all recording formats
	for _, f := range rec.Recording.Formats {
		// Check format is safe
		format := f.Type
		if err := assertFsSafe(format); err != nil {
			return err
		}

		// Base path for a recording format
		fmtSrcBase := filepath.Join(src, format)
		fmtDstBase := filepath.Join(dst, format)

		// Make sure the destination base path exists
		if err := unsafeAssertFsPath(fmtDstBase); err != nil {
			return err
		}

		recSrc := filepath.Join(fmtSrcBase, recID)
		recDst := filepath.Join(fmtDstBase, recID)
		if err := unsafeMoveFiles(recSrc, recDst); err != nil {
			return err
		}

		log.Debug().
			Str("recID", recID).Str("format", format).
			Str("src", src).Str("dst", dst).
			Str("recSrc", recSrc).Str("recDst", recDst).
			Msg("moved recording files")
	}

	return nil
}

// PublishRecording will move the recording to the published path.
func (s *RecordingsStorage) PublishRecording(
	rec *RecordingState,
) error {
	return s.saveMoveRecording(rec, s.UnpublishedPath, s.PublishedPath)
}

// UnpublishRecording will move the recording to the unpublished path
func (s *RecordingsStorage) UnpublishRecording(
	rec *RecordingState,
) error {
	return s.saveMoveRecording(rec, s.PublishedPath, s.UnpublishedPath)
}

// ImportRecording will import the recording files to either
// the unpublished or published path.
func (s *RecordingsStorage) ImportRecording(rec *RecordingState) error {
	dst := s.PublishedPath
	if !rec.Recording.Published {
		dst = s.UnpublishedPath
	}
	return s.saveMoveRecording(rec, s.InboxPath, dst)
}
