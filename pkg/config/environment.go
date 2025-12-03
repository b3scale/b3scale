package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/rs/zerolog/log"
)

// Well Known Environment Keys
const (
	EnvBBBConfig = "BBB_CONFIG"

	EnvDbURL      = "B3SCALE_DB_URL"
	EnvDbPoolSize = "B3SCALE_DB_POOL_SIZE"

	EnvLogLevel  = "B3SCALE_LOG_LEVEL"
	EnvLogFormat = "B3SCALE_LOG_FORMAT"

	EnvListenHTTP   = "B3SCALE_LISTEN_HTTP"
	EnvReverseProxy = "B3SCALE_REVERSE_PROXY_MODE"
	EnvLoadFactor   = "B3SCALE_LOAD_FACTOR"

	EnvJWTSecret      = "B3SCALE_API_JWT_SECRET"
	EnvAPIURL         = "B3SCALE_API_URL"
	EnvAPIAccessToken = "B3SCALE_API_ACCESS_TOKEN"

	EnvRecordingsInboxPath         = "B3SCALE_RECORDINGS_INBOX_PATH"
	EnvRecordingsPublishedPath     = "B3SCALE_RECORDINGS_PUBLISHED_PATH"
	EnvRecordingsUnpublishedPath   = "B3SCALE_RECORDINGS_UNPUBLISHED_PATH"
	EnvRecordingsPlaybackHost      = "B3SCALE_RECORDINGS_PLAYBACK_HOST"
	EnvRecordingsDefaultVisibility = "B3SCALE_RECORDINGS_DEFAULT_VISIBILITY"

	EnvHTTPRequestTimeout    = "B3SCALE_HTTP_REQUEST_TIMEOUT"
	EnvHTTPReadHeaderTimeout = "B3SCALE_HTTP_READ_HEADER_TIMEOUT"
	EnvHTTPWriteTimeout      = "B3SCALE_HTTP_WRITE_TIMEOUT"
	EnvHTTPIdleTimeout       = "B3SCALE_HTTP_IDLE_TIMEOUT"
)

// Defaults
const (
	EnvBBBConfigDefault = "/etc/bigbluebutton/bbb-web.properties"

	EnvDbPoolSizeDefault = "128"
	EnvDbURLDefault      = "postgres://postgres:postgres@localhost:5432/b3scale"

	EnvLogLevelDefault  = "info"
	EnvLogFormatDefault = "structured"

	EnvReverseProxyDefault = "false"
	EnvLoadFactorDefault   = "1.0"

	EnvRecordingsDefaultVisibilityDefault = "published"

	EnvListenHTTPDefault = "127.0.0.1:42353" // :B3S

	// HTTP timeout defaults (in seconds)
	EnvHTTPRequestTimeoutDefault    = "60"
	EnvHTTPReadHeaderTimeoutDefault = "5"
	EnvHTTPWriteTimeoutDefault      = "60"
	EnvHTTPIdleTimeoutDefault       = "120"
)

// LoadEnv loads the environment from a file and
// updates the os.Environment
func LoadEnv(envfiles []string) {
	for _, filename := range envfiles {
		loadEnvFile(filename)
	}
}

// Internal load a single env file
func loadEnvFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		// We could not open the file - so we ignore this.
		return
	}
	defer f.Close()

	fmt.Println("using environment from:", filename)

	// Read lines and set env
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "#") {
			continue // comment
		}
		tokens := strings.SplitN(l, "=", 2)
		if len(tokens) != 2 {
			continue // invalid
		}
		k := strings.Trim(strings.TrimSpace(tokens[0]), "\"'")
		v := strings.Trim(strings.TrimSpace(tokens[1]), "\"'")
		os.Setenv(k, v)
	}
}

// EnvOpt gets a configuration from the environment
// with a default fallback.
func EnvOpt(key, fallback string) string {
	value, ok := GetEnvOpt(key)
	if !ok {
		return fallback
	}
	return value
}

// GetEnvOpt gets a configuration from the environment,
// but will fail if the variable is not present.
func GetEnvOpt(key string) (string, bool) {
	value := os.Getenv(key)
	if value == "" {
		return "", false
	}
	return value, true
}

// MustEnv gets a configuration from the environment
// and will panic if the variable is empty.
func MustEnv(key string) string {
	value, ok := GetEnvOpt(key)
	if !ok {
		err := fmt.Errorf("missing environment configuration: %s", key)
		panic(err)
	}
	return value
}

// IsEnabled returns true if the input is trueish
func IsEnabled(value string) bool {
	value = strings.ToLower(value)
	switch value {
	case "yes":
		return true
	case "true":
		return true
	case "1":
		return true
	}
	return false
}

// GetLoadFactor retrievs the load factor
// from the environment.
func GetLoadFactor() float64 {
	val := EnvOpt(EnvLoadFactor, EnvLoadFactorDefault)
	factor, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Error().Err(err).Msg("invalid value for " + EnvLoadFactor)
		return 1.0
	}
	return factor
}

// DomainOf returns the domain name (with TLD) of the given
// address or URL.
// FIXME: This feels out of place here.
func DomainOf(addr string) string {
	u, err := url.Parse(addr)
	if err != nil {
		panic(err)
	}
	host := u.Hostname()
	if host == "" {
		host = addr
	}
	tokens := strings.Split(host, ".")
	if len(tokens) < 2 {
		return tokens[0]
	}
	domain := tokens[len(tokens)-2] + "." + tokens[len(tokens)-1]
	return domain
}

// Internal: Get default visibility from environemnt and parse
// into a RecordingVisiblity enum.
func envGetRecordingsDefaultVisibility() (bbb.RecordingVisibility, error) {
	repr := EnvOpt(
		EnvRecordingsDefaultVisibility,
		EnvRecordingsDefaultVisibilityDefault)
	return bbb.ParseRecordingVisibility(repr)
}

// GetRecordingsDefaultVisibility returns the parsed default
// visibility from the environment.
//
// This function will never panic: CheckEnv will ensure that
// the configured value is valid. Make sure CheckEnv is invoked
// prior to using this function.
func GetRecordingsDefaultVisibility() bbb.RecordingVisibility {
	v, _ := envGetRecordingsDefaultVisibility()
	return v
}

// GetRecordingsPublishedPath returns the configured
// published path.
func GetRecordingsPublishedPath() string {
	return os.Getenv(EnvRecordingsPublishedPath)
}

// GetRecordingsUnpublishedPath returns the configured
// path to unpublished recordings.
func GetRecordingsUnpublishedPath() string {
	return os.Getenv(EnvRecordingsUnpublishedPath)
}

// GetRecordingsInboxPath returns the configured inbox
// path. If the environment variable is not set,
// either the published or unpublished path will be returned
// depending on the default visibility.
func GetRecordingsInboxPath() string {
	v := GetRecordingsDefaultVisibility()
	p := os.Getenv(EnvRecordingsInboxPath)
	if p == "" && v == bbb.RecordingVisibilityUnpublished {
		return GetRecordingsUnpublishedPath()
	}
	if p == "" {
		return GetRecordingsPublishedPath()
	}
	return p
}

// requireEnv checks if env vars are set and returns missing ones.
func requireEnv(keys ...string) []string {
	var missing []string
	for _, key := range keys {
		if _, ok := GetEnvOpt(key); !ok {
			missing = append(missing, key)
		}
	}
	return missing
}

// checkAPIConfig checks API configuration and logs settings.
func checkAPIConfig() ([]string, error) {
	missing := requireEnv(EnvAPIURL, EnvJWTSecret)
	return missing, nil
}

// checkRecordingsConfig checks recordings configuration and logs settings.
func checkRecordingsConfig() ([]string, error) {
	vis, err := envGetRecordingsDefaultVisibility()
	if err != nil {
		log.Error().Err(err).Msg("invalid recordings visibility config")
		return nil, err
	}

	inPath, _ := GetEnvOpt(EnvRecordingsInboxPath)
	pubPath, hasPubPath := GetEnvOpt(EnvRecordingsPublishedPath)
	unpubPath, hasUnpubPath := GetEnvOpt(EnvRecordingsUnpublishedPath)

	recEnabled := hasPubPath || hasUnpubPath

	var missing []string
	if !hasPubPath {
		missing = append(missing, EnvRecordingsPublishedPath)
	}
	if !hasUnpubPath {
		missing = append(missing, EnvRecordingsUnpublishedPath)
	}

	log.Info().
		Bool("recordings_enabled", recEnabled).
		Str("inbox_path", inPath).
		Str("published_path", pubPath).
		Str("unpublished_path", unpubPath).
		Str("default_visibility", vis.String()).
		Msg("recordings settings")

	return missing, nil
}

// checkHTTPConfig checks HTTP configuration and logs settings.
func checkHTTPConfig() ([]string, error) {
	listen := EnvOpt(EnvListenHTTP, EnvListenHTTPDefault)
	log.Info().Str("listen", listen).Msg("http listen address")

	log.Info().
		Float64("request_timeout", GetHTTPRequestTimeout().Seconds()).
		Float64("read_header_timeout", GetHTTPReadHeaderTimeout().Seconds()).
		Float64("write_timeout", GetHTTPWriteTimeout().Seconds()).
		Float64("idle_timeout", GetHTTPIdleTimeout().Seconds()).
		Msg("http timeout settings (in seconds)")

	return nil, nil
}

// CheckEnv checks if the environment is configured
func CheckEnv() error {
	var missing []string

	checks := []func() ([]string, error){
		checkHTTPConfig,
		checkAPIConfig,
		checkRecordingsConfig,
	}

	for _, check := range checks {
		m, err := check()
		if err != nil {
			return err
		}
		missing = append(missing, m...)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing environment variables: %s",
			strings.Join(missing, ", "))
	}

	return nil
}

// getEnvTimeoutSec parses an environment variable as seconds (int)
// and returns a time.Duration. Logs error and uses default if invalid.
func getEnvTimeoutSec(key, fallback string) time.Duration {
	val := EnvOpt(key, fallback)
	seconds, err := strconv.Atoi(val)
	if err != nil {
		log.Error().Err(err).Str("key", key).Str("value", val).
			Msg("invalid timeout value (expected seconds), using default")
		seconds, _ = strconv.Atoi(fallback)
	}
	return time.Duration(seconds) * time.Second
}

// GetHTTPRequestTimeout returns the HTTP request timeout.
func GetHTTPRequestTimeout() time.Duration {
	return getEnvTimeoutSec(EnvHTTPRequestTimeout, EnvHTTPRequestTimeoutDefault)
}

// GetHTTPReadHeaderTimeout returns the HTTP read header timeout.
func GetHTTPReadHeaderTimeout() time.Duration {
	return getEnvTimeoutSec(EnvHTTPReadHeaderTimeout, EnvHTTPReadHeaderTimeoutDefault)
}

// GetHTTPWriteTimeout returns the HTTP write timeout.
func GetHTTPWriteTimeout() time.Duration {
	return getEnvTimeoutSec(EnvHTTPWriteTimeout, EnvHTTPWriteTimeoutDefault)
}

// GetHTTPIdleTimeout returns the HTTP idle timeout.
func GetHTTPIdleTimeout() time.Duration {
	return getEnvTimeoutSec(EnvHTTPIdleTimeout, EnvHTTPIdleTimeoutDefault)
}
