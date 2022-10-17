package openapi

// Endpoints combines path mappings
func Endpoints(schemas ...map[string]Path) map[string]Path {
	combined := map[string]Path{}
	// Merge endpoints into a single schema
	for _, endpoints := range schemas {
		for path, operations := range endpoints {
			combined[path] = operations
		}
	}
	return combined
}
