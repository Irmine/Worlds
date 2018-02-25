package chunks

// Viewer is a viewer of a chunk.
type Viewer interface {
	GetRuntimeId() uint64
}
