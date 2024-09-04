package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// Well Known Environment Keys
const (
	EnvDbURL                     = "B3SCALE_DB_URL"
	EnvDbPoolSize                = "B3SCALE_DB_POOL_SIZE"
	EnvLogLevel                  = "B3SCALE_LOG_LEVEL"
	EnvLogFormat                 = "B3SCALE_LOG_FORMAT"
	EnvListenHTTP                = "B3SCALE_LISTEN_HTTP"
	EnvReverseProxy              = "B3SCALE_REVERSE_PROXY_MODE"
	EnvLoadFactor                = "B3SCALE_LOAD_FACTOR"
	EnvJWTSecret                 = "B3SCALE_API_JWT_SECRET"
	EnvAPIURL                    = "B3SCALE_API_URL"
	EnvAPIAccessToken            = "B3SCALE_API_ACCESS_TOKEN"
	EnvBBBConfig                 = "BBB_CONFIG"
	EnvRecordingsPublishedPath   = "B3SCALE_RECORDINGS_PUBLISHED_PATH"
	EnvRecordingsUnpublishedPath = "B3SCALE_RECORDINGS_UNPUBLISHED_PATH"
	EnvRecordingsPlaybackHost    = "B3SCALE_RECORDINGS_PLAYBACK_HOST"
)

// Defaults
const (
	EnvDbPoolSizeDefault   = "128"
	EnvDbURLDefault        = "postgres://postgres:postgres@localhost:5432/b3scale"
	EnvLogLevelDefault     = "info"
	EnvLogFormatDefault    = "structured"
	EnvListenHTTPDefault   = "127.0.0.1:42353" // :B3S
	EnvReverseProxyDefault = "false"
	EnvBBBConfigDefault    = "/etc/bigbluebutton/bbb-web.properties"
	EnvLoadFactorDefault   = "1.0"
)

// LoadEnv loads the environment from a file and
// updates the os.Environment
func LoadEnv(envfiles []string) {
	for _, filename := range envfiles {
		loadEnvFile(filename)
	}
}

// CheckEnv checks if the environment is configured
func CheckEnv() error {
	missing := []string{}

	if _, ok := GetEnvOpt(EnvAPIURL); !ok {
		missing = append(missing, EnvAPIURL)
	}

	if _, ok := GetEnvOpt(EnvJWTSecret); !ok {
		missing = append(missing, EnvJWTSecret)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing environment variables: %s",
			strings.Join(missing, ", "))
	}

	return nil
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
