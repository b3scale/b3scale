package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

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
)

// Defaults
const (
	EnvBBBConfigDefault = "/etc/bigbluebutton/bbb-web.properties"

	EnvDbPoolSizeDefault = "128"
	EnvDbURLDefault      = "postgres://postgres:postgres@localhost:5432/b3scale"

	EnvLogLevelDefault  = "info"
	EnvLogFormatDefault = "structured"

	EnvListenHTTPDefault   = "127.0.0.1:42353" // :B3S
	EnvReverseProxyDefault = "false"
	EnvLoadFactorDefault   = "1.0"

	EnvRecordingsDefaultVisibilityDefault = "published"
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

// CheckEnv checks if the environment is configured
func CheckEnv() error {
	missing := []string{}

	// API and Secret
	if _, ok := GetEnvOpt(EnvAPIURL); !ok {
		missing = append(missing, EnvAPIURL)
	}

	if _, ok := GetEnvOpt(EnvJWTSecret); !ok {
		missing = append(missing, EnvJWTSecret)
	}

	// Recordings Default Visibility
	vis, err := envGetRecordingsDefaultVisibility()
	if err != nil {
		return err
	}

	// Recordings paths
	// In case the Published Path is configured, check that
	// the configuration is complete.
	recEnabled := false
	inPath, hasInPath := GetEnvOpt(EnvRecordingsInboxPath)
	pubPath, hasPubPath := GetEnvOpt(EnvRecordingsPublishedPath)
	unpubPath, hasUnpubPath := GetEnvOpt(EnvRecordingsUnpublishedPath)
	if hasInPath || hasPubPath || hasUnpubPath {
		if !hasInPath {
			missing = append(missing, EnvRecordingsInboxPath)
		}
		if !hasPubPath {
			missing = append(missing, EnvRecordingsPublishedPath)
		}
		if !hasUnpubPath {
			missing = append(missing, EnvRecordingsUnpublishedPath)
		}
		recEnabled = true
	}

	// Log recording settings
	log.Info().
		Bool("recordings_enabled", recEnabled).
		Str("inbox_path", inPath).
		Str("published_path", pubPath).
		Str("unpublished_path", unpubPath).
		Str("default_visibility", vis.String()).
		Msg("recordings settings")

	if len(missing) > 0 {
		return fmt.Errorf("missing environment variables: %s",
			strings.Join(missing, ", "))
	}

	return nil
}
