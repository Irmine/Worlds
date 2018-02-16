package chunks

// SubChunk is a 16x16x16 segment in a chunk.
// A chunk contains 16 sub chunks.
type SubChunk struct {
	BlockIds   []byte
	BlockData  []byte
	BlockLight []byte
	SkyLight   []byte
}

// NewSubChunk returns a new sub chunk.
func NewSubChunk() *SubChunk {
	return &SubChunk{make([]byte, 4096), make([]byte, 2048), make([]byte, 2048), make([]byte, 2048)}
}

// IsAllAir checks if the sub chunk is completely made up out of air.
func (subChunk *SubChunk) IsAllAir() bool {
	return string(subChunk.BlockIds) == string(make([]byte, 4096))
}

// GetIdIndex returns the block ID index of the given position.
func (subChunk *SubChunk) GetIdIndex(x, y, z int) int {
	return (x << 8) | (z << 4) | y
}

// GetDataIndex returns the block data index of the given position.
func (subChunk *SubChunk) GetDataIndex(x, y, z int) int {
	return (x << 7) + (z << 3) + (y >> 1)
}

// GetBlockId returns the block ID of the block on the given position.
func (subChunk *SubChunk) GetBlockId(x, y, z int) byte {
	return subChunk.BlockIds[subChunk.GetIdIndex(x, y, z)]
}

// SetBlockId sets the block ID at the given position to the given ID.
func (subChunk *SubChunk) SetBlockId(x, y, z int, id byte) {
	subChunk.BlockIds[subChunk.GetIdIndex(x, y, z)] = id
}

// GetBlockLight returns the block light on the given position.
func (subChunk *SubChunk) GetBlockLight(x, y, z int) byte {
	var data = subChunk.BlockLight[subChunk.GetDataIndex(x, y, z)]
	if (y & 0x01) == 0 {
		return data & 0x0f
	}
	return data >> 4
}

// SetBlockLight sets the block light on the given position.
func (subChunk *SubChunk) SetBlockLight(x, y, z int, light byte) {
	var i = subChunk.GetDataIndex(x, y, z)
	var d = subChunk.BlockLight[i]
	if (y & 0x01) == 0 {
		subChunk.BlockLight[i] = (d & 0xf0) | (light & 0x0f)
		return
	}
	subChunk.BlockLight[i] = ((light & 0x0f) << 4) | (d & 0x0f)
}

// GetSkySlight returns the skylight at the given position.
func (subChunk *SubChunk) GetSkyLight(x, y, z int) byte {
	var data = subChunk.SkyLight[subChunk.GetDataIndex(x, y, z)]
	if (y & 0x01) == 0 {
		return data & 0x0f
	}
	return data >> 4
}

// SetSkyLight sets the skylight at the given position.
func (subChunk *SubChunk) SetSkyLight(x, y, z int, light byte) {
	var i = subChunk.GetDataIndex(x, y, z)
	var d = subChunk.SkyLight[i]
	if (y & 0x01) == 0 {
		subChunk.SkyLight[i] = (d & 0xf0) | (light & 0x0f)
		return
	}
	subChunk.SkyLight[i] = ((light & 0x0f) << 4) | (d & 0x0f)
}

// GetBlockData returns the block data at the given position.
func (subChunk *SubChunk) GetBlockData(x, y, z int) byte {
	var data = subChunk.BlockData[subChunk.GetDataIndex(x, y, z)]
	if (y & 0x01) == 0 {
		return data & 0x0f
	}
	return data >> 4
}

// SetBlockData sets the block data at the given position.
func (subChunk *SubChunk) SetBlockData(x, y, z int, data byte) {
	var i = subChunk.GetDataIndex(x, y, z)
	var d = subChunk.BlockData[i]
	if (y & 0x01) == 0 {
		subChunk.BlockData[i] = (d & 0xf0) | (data & 0x0f)
		return
	}
	subChunk.BlockData[i] = ((data & 0x0f) << 4) | (d & 0x0f)
}

// GetHighestBlockId returns the block ID of the highest block in the given column.
func (subChunk *SubChunk) GetHighestBlockId(x, z int) byte {
	for y := 15; y >= 0; y-- {
		id := subChunk.GetBlockId(x, y, z)
		if id != 0 {
			return id
		}
	}
	return 0
}

// GetHighestBlockData returns the block data of the highest block in the given column.
func (subChunk *SubChunk) GetHighestBlockData(x, z int) byte {
	for y := 15; y >= 0; y-- {
		return subChunk.GetBlockData(x, y, z)
	}

	return 0
}

// GetHighestBlockY returns the Y value of the highest block in the given column.
func (subChunk *SubChunk) GetHighestBlockY(x, z int) int16 {
	for y := 15; y >= 0; y-- {
		if subChunk.GetBlockId(x, y, z) != 0 {
			return int16(y)
		}
	}

	return 0
}

// ToBinary returns the binary representation of the sub chunk used for network sending.
func (subChunk *SubChunk) ToBinary() []byte {
	var bytes = []byte{00}
	bytes = append(bytes, subChunk.BlockIds...)
	bytes = append(bytes, subChunk.BlockData...)
	return bytes
}
