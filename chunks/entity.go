package chunks

import (
	"github.com/golang/geo/r3"
	"github.com/irmine/gonbt"
)

type ChunkEntity interface {
	GetRuntimeId() uint64
	IsClosed() bool
	GetEntityId() uint32
	GetPosition() r3.Vector
	GetSaveData() *gonbt.Compound
}
