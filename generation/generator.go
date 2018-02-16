package generation

import (
	"github.com/irmine/worlds/chunks"
)

// Generators generate segments of worlds.
type Generator interface {
	GetName() string
	GenerateNewChunk(x, z int32) *chunks.Chunk
}
