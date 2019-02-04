package defaults

import (
	"github.com/irmine/worlds/chunks"
)

type Flat struct {
}

func NewFlatGenerator() Flat {
	return Flat{}
}

func (f Flat) GetName() string {
	return "Flat"
}

func (f Flat) GenerateNewChunk(x, z int32) *chunks.Chunk {
	var chunk = chunks.New(x, z)
	var y int
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			y = 0
			chunk.SetBlockId(x, y, z, 7)
			y++
			for y2 := y; y2 < 4; y2++ {
				chunk.SetBlockId(x, y2, z, 3)
				y++
			}
			chunk.SetBlockId(x, y, z, 2)
		}
	}
	chunk.RecalculateHeightMap()
	return chunk
}