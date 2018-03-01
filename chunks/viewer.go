package chunks

import "github.com/irmine/gomine/utils"

// Viewer is a viewer of a chunk.
type Viewer interface {
	GetUUID() utils.UUID
	GetXUID() string
}
