package utils

import (
	"github.com/golang/geo/r3"
	"github.com/irmine/worlds/blocks"
)

func VectorToPosition(vector r3.Vector) blocks.Position {
	return blocks.Position{X: int32(vector.X), Z: int32(vector.Z), Y: uint32(vector.Y)}
}

func PositionToVector(position blocks.Position) r3.Vector {
	return r3.Vector{X: float64(position.X), Y: float64(position.Y), Z: float64(position.Z)}
}