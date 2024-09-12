package config

/*
 Userdir config file store: Retrieve and store data in the
 config directory.
*/

import (
	"os"
	"path"
	"regexp"
)

var (
	// ReMatchUnsafe matches everything not a-z, A-Z, 0-9
	// and '.' from a string.
	ReMatchUnsafe = regexp.MustCompile(`[^a-zA-Z0-9.]`)

	// ReMatchUnderscoreSeq matches underscore sequences
	ReMatchUnderscoreSeq = regexp.MustCompile(`__+`)
)

// SafeFilename creates an urlsafe filename by stripping
// unsafe characters.
func SafeFilename(f string) string {
	f = ReMatchUnsafe.ReplaceAllString(f, "_")
	f = ReMatchUnderscoreSeq.ReplaceAllString(f, "_")
	return f
}

// UserDirPath joins the file name with the
// full b3scale config path
func UserDirPath(suffix string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configPath := path.Join(base, "b3scale", SafeFilename(suffix))
	return configPath, nil
}

// UserDirPut save a file in the b3scale config directory
func UserDirPut(filename string, data []byte) error {
	configPath, err := UserDirPath("")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return err
	}
	fullPath := path.Join(configPath, SafeFilename(filename))
	return os.WriteFile(fullPath, data, 0600)
}

// UserDirFilename gets the full path to a filename in the userdir
func UserDirFilename(filename string) (string, error) {
	configPath, err := UserDirPath("")
	if err != nil {
		return "", err
	}
	fullPath := path.Join(configPath, SafeFilename(filename))
	return fullPath, err
}

// UserDirGet retrievs content from a file in the
// b3scale user config directory
func UserDirGet(filename string) ([]byte, error) {
	configPath, err := UserDirPath("")
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return nil, err
	}
	fullPath := path.Join(configPath, SafeFilename(filename))
	return os.ReadFile(fullPath)
}

// UserDirGetString retrievs a string from a file
// in the b3scale user config directory
func UserDirGetString(filename string) (string, error) {
	data, err := UserDirGet(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
