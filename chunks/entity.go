package chunks

import (
	"github.com/golang/geo/r3"
	"github.com/irmine/gonbt"
)

type ChunkEntity interface {
	GetRuntimeId() uint64
	SetRuntimeId(id uint64)
	IsClosed() bool
	GetEntityType() uint32
	Close()
	GetPosition() r3.Vector
	SetPosition(r3.Vector) error
	GetNBT() *gonbt.Compound
	SetNBT(*gonbt.Compound)
	SetDimension(interface {
		GetChunk(int32, int32) (*Chunk, bool)
	})
	SetLevel(interface {
		DimensionExists(string) bool
	})
	SpawnToAll()
	Tick()
}
