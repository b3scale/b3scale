package cluster

const (
	// BackendLabel marks a backend label
	BackendLabel = iota
	// FrontendLabel marks a frontend label
	FrontendLabel
	// TagLabel marks a tag label
	TagLabel
)

// A Label contains routing information
type Label struct {
	Key   int
	value interface{}
}

// Labels are a stack of routing information
type Labels []*Label

// Append a label
func (labels Labels) add(label *Label) {

}

// Remove a label
func (labels Labels) remove(label *Label) {
}

// Find a label of a given type
func (labels Labels) findKey(key int) *Label {
	for _, l := range labels {
		if l.Key == key {
			return l
		}
	}
	return nil
}

// Find and remove
func (labels Labels) pop(key int) *Label {
	// Find a backend label
	label := labels.findKey(BackendLabel)
	if label == nil {
		return nil
	}
	labels.remove(label)
	return label
}

// PopBackend removes a backend label
func (labels Labels) PopBackend() *Backend {
	label := labels.pop(BackendLabel)
	if label == nil {
		return nil
	}
	return label.value.(*Backend)
}
