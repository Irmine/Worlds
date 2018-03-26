package chunks

import (
	"github.com/google/uuid"
)

// Viewer is a viewer of a chunk.
type Viewer interface {
	GetUUID() uuid.UUID
	GetXUID() string
}
