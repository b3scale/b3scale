package config

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

/*
 Read a BBB .properties file
*/

var rePropertyRef *regexp.Regexp = regexp.MustCompile(`\${(.*)}`)

// Properties is a map of BBB properties.
// The map stores the raw data. Retriving values should
// be done through the accessor which will resolve refs.
type Properties map[string]string

// Get retrievs a value from the properties map and
// will resolve a reference.
func (p Properties) Get(key string) (string, bool) {
	var (
		val string
		ok  bool
	)
	val, ok = p[key]
	if !ok {
		return "", false
	}

	// Resolve ref if any
	val = rePropertyRef.ReplaceAllStringFunc(val, func(ref string) string {
		var rval string
		rkey := ref[2 : len(ref)-1]
		rval, ok = p[rkey]
		return rval
	})

	return val, ok
}

// ReadPropertiesFile consumes a BBB properties file
func ReadPropertiesFile(filename string) (Properties, error) {
	// This is pretty much linewise key=value
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	props := make(Properties)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue // Skip comments
		}
		if !strings.ContainsRune(line, '=') {
			continue // Skip non assignments
		}

		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) < 2 {
			continue // Huh.
		}

		k := strings.TrimSpace(tokens[0])
		v := strings.TrimSpace(tokens[1])

		props[k] = v
	}

	return props, nil
}
