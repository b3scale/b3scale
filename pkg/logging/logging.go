package logging

// B3Scale logging configuration

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Options for logging. Get them from the env or a
// config or something.
type Options struct {
	Level  string
	Format string
}

// Parse the log level string. This can either be a
// numeric or a string value. See documentation
// of zerolog for possible values.
func parseLogLevel(level string) (zerolog.Level, error) {
	var (
		loglevel zerolog.Level
		err      error
	)
	value, err := strconv.ParseInt(level, 10, 8)
	if err != nil {
		// Try string parsing
		loglevel, err = zerolog.ParseLevel(level)
		if err != nil {
			return zerolog.NoLevel, err
		}
	} else {
		loglevel = zerolog.Level(value)
	}
	// Check level
	if loglevel < zerolog.TraceLevel || loglevel > 5 {
		return zerolog.NoLevel, fmt.Errorf("Invalid error level, out of range")
	}

	return loglevel, nil
}

// Setup configures the log level and sets a
// console write unless not confgured otherwise
func Setup(opts *Options) error {
	loglevel, err := parseLogLevel(opts.Level)
	if err != nil {
		return err
	}

	// Setup logging
	zerolog.SetGlobalLevel(loglevel)

	if opts.Format != "structured" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	return nil
}
