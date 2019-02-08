package chunks

import (
	"github.com/google/uuid"
	"github.com/irmine/worlds/blocks"
)

// Viewer is a viewer of a chunk.
type Viewer interface {
	GetUUID() uuid.UUID
	GetXUID() string
	SendUpdateBlock(position blocks.Position, blockRuntimeId, dataLayerId uint32)
}
