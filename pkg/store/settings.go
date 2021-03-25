package store

// Settings hold per front or backend runtime configuration.
// Variables can be accessed during request routing and
// handling in middlewares.
type Settings map[string]interface{}

// SettingsValue is a generic settings value
type SettingsValue interface{}

// Get retrievs a value with a fallback
func (s Settings) Get(key string, fallback SettingsValue) SettingsValue {
	val, ok := s[key]
	if !ok {
		return fallback
	}
	return val
}

// GetString returns the settings value as string
func (s Settings) GetString(key, fallback string) string {
	val, ok := s.Get(key, fallback).(string)
	if !ok {
		return fallback
	}
	return val
}

// GetInt returns the settings value as integer
func (s Settings) GetInt(key string, fallback int) int {
	val, ok := s.Get(key, fallback).(int)
	if !ok {
		return fallback
	}
	return val
}

// GetBool returns the settings value as boolean
func (s Settings) GetBool(key string, fallback bool) bool {
	val, ok := s.Get(key, fallback).(bool)
	if !ok {
		return fallback
	}
	return val
}

// Set a value for a key in settings
func (s Settings) Set(key string, value SettingsValue) {
	if value == nil {
		delete(s, key)
		return
	}
	s[key] = value
}
