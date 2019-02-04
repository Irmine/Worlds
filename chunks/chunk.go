package chunks

import (
	"errors"
	"github.com/google/uuid"
	"github.com/irmine/binutils"
	"github.com/irmine/gonbt"
	"sync"
)

// Chunk is a segment of the world, holding blocks, block data, biomes, etc.
type Chunk struct {
	X, Z             int32
	LightPopulated   bool
	TerrainPopulated bool
	Biomes           []byte
	HeightMap        []int16

	InhabitedTime int64
	LastUpdate    int64

	*sync.RWMutex
	viewers   map[uuid.UUID]Viewer
	entities  map[uint64]ChunkEntity
	blockNBT  map[int]*gonbt.Compound
	subChunks map[byte]*SubChunk
}

// New returns a new chunk with the given X and Z.
func New(x, z int32) *Chunk {
	return &Chunk{x, z,
		true,
		true,
		make([]byte, 256),
		make([]int16, 256),
		0,
		0,
		&sync.RWMutex{},
		make(map[uuid.UUID]Viewer),
		make(map[uint64]ChunkEntity),
		make(map[int]*gonbt.Compound),
		make(map[byte]*SubChunk),
	}
}

// GetViewers returns all viewers of the chunk.
// Viewers are all players that have the chunk within their view distance.
func (chunk *Chunk) GetViewers() map[uuid.UUID]Viewer {
	return chunk.viewers
}

// AddViewer adds a viewer of the chunk.
func (chunk *Chunk) AddViewer(player Viewer) {
	chunk.Lock()
	chunk.viewers[player.GetUUID()] = player
	chunk.Unlock()
}

// RemoveViewer removes a viewer from the chunk.
func (chunk *Chunk) RemoveViewer(player Viewer) {
	chunk.Lock()
	delete(chunk.viewers, player.GetUUID())
	chunk.Unlock()
}

// GetBiome returns the biome at the given column.
func (chunk *Chunk) GetBiome(x, z int) byte {
	return chunk.Biomes[chunk.GetBiomeIndex(x, z)]
}

// SetBiome sets the biome at the given column.
func (chunk *Chunk) SetBiome(x, z int, biome byte) {
	chunk.Biomes[chunk.GetBiomeIndex(x, z)] = biome
}

// AddEntity adds a new entity to the chunk.
func (chunk *Chunk) AddEntity(entity ChunkEntity) error {
	if entity.IsClosed() {
		return errors.New("cannot add closed entity to chunk")
	}
	chunk.Lock()
	chunk.entities[entity.GetRuntimeId()] = entity
	chunk.Unlock()
	return nil
}

// RemoveEntity removes an entity with the given runtimeId from the chunk.
func (chunk *Chunk) RemoveEntity(runtimeId uint64) {
	chunk.Lock()
	delete(chunk.entities, runtimeId)
	chunk.Unlock()
}

// GetEntities returns all entities of the chunk.
func (chunk *Chunk) GetEntities() map[uint64]ChunkEntity {
	return chunk.entities
}

// SetBlockNBTAt sets the given compound at the given position.
func (chunk *Chunk) SetBlockNBTAt(x, y, z int, nbt *gonbt.Compound) {
	chunk.Lock()
	if nbt == nil {
		delete(chunk.blockNBT, GetBlockNBTIndex(x, y, z))
	} else {
		chunk.blockNBT[GetBlockNBTIndex(x, y, z)] = nbt
	}
	chunk.Unlock()
}

// RemoveBlockNBTAt removes the block NBT at the given position.
func (chunk *Chunk) RemoveBlockNBTAt(x, y, z int) {
	chunk.Lock()
	delete(chunk.blockNBT, GetBlockNBTIndex(x, y, z))
	chunk.Unlock()
}

// BlockNBTExistsAt checks if any block NBT exists at the given position.
func (chunk *Chunk) BlockNBTExistsAt(x, y, z int) bool {
	chunk.RLock()
	var _, ok = chunk.blockNBT[GetBlockNBTIndex(x, y, z)]
	chunk.RUnlock()
	return ok
}

// GetBlockNBTAt returns the block NBT at the given position.
// Returns a bool if any block NBT was found at that position
func (chunk *Chunk) GetBlockNBTAt(x, y, z int) (*gonbt.Compound, bool) {
	chunk.RLock()
	var c, ok = chunk.blockNBT[GetBlockNBTIndex(x, y, z)]
	chunk.RUnlock()
	return c, ok
}

// GetBiomeIndex returns the biome index of a column in a chunk.
func (chunk *Chunk) GetBiomeIndex(x, z int) int {
	return (x << 4) | z
}

// GetIndex returns the index of a block position in a chunk.
func (chunk *Chunk) GetIndex(x, y, z int) int {
	return (x << 12) | (z << 8) | y
}

// GetHeightMapIndex returns the index of a height map position in a chunk.
func (chunk *Chunk) GetHeightMapIndex(x, z int) int {
	return (z << 4) | x
}

// SetBlockId sets the given block ID at the given position.
func (chunk *Chunk) SetBlockId(x, y, z int, blockId byte) {
	chunk.GetSubChunk(byte(y>>4)).SetBlockId(x, y&15, z, blockId)
}

// GetBlockId returns the block ID of a block at the given position.
func (chunk *Chunk) GetBlockId(x, y, z int) byte {
	return chunk.GetSubChunk(byte(y>>4)).GetBlockId(x, y&15, z)
}

// SetBlockData sets the block data of a block at the given position.
func (chunk *Chunk) SetBlockData(x, y, z int, data byte) {
	chunk.GetSubChunk(byte(y>>4)).SetBlockData(x, y&15, z, data)
}

// GetBlockData returns the block data of a block at the given position.
func (chunk *Chunk) GetBlockData(x, y, z int) byte {
	return chunk.GetSubChunk(byte(y>>4)).GetBlockData(x, y&15, z)
}

// SetBlockLight sets the block light on a position in this chunk.
func (chunk *Chunk) SetBlockLight(x, y, z int, level byte) {
	chunk.GetSubChunk(byte(y>>4)).SetBlockLight(x, y&15, z, level)
}

// GetBlockLight returns the block light on a position in this chunk.
func (chunk *Chunk) GetBlockLight(x, y, z int) byte {
	return chunk.GetSubChunk(byte(y>>4)).GetBlockLight(x, y&15, z)
}

// SetSkyLight sets the sky light on a position in this chunk.
func (chunk *Chunk) SetSkyLight(x, y, z int, level byte) {
	chunk.GetSubChunk(byte(y>>4)).SetSkyLight(x, y&15, z, level)
}

// GetSkyLight returns the sky light on a position in this chunk.
func (chunk *Chunk) GetSkyLight(x, y, z int) byte {
	return chunk.GetSubChunk(byte(y>>4)).GetSkyLight(x, y&15, z)
}

// SetSubChunk sets a SubChunk on a position in this chunk.
func (chunk *Chunk) SetSubChunk(y byte, subChunk *SubChunk) {
	chunk.Lock()
	chunk.subChunks[y] = subChunk
	chunk.Unlock()
}

// GetSubChunk returns a SubChunk on a given height index in this chunk.
func (chunk *Chunk) GetSubChunk(y byte) *SubChunk {
	chunk.RLock()
	if sub, ok := chunk.subChunks[y]; ok {
		chunk.RUnlock()
		return sub
	}
	chunk.RUnlock()
	chunk.Lock()
	defer chunk.Unlock()
	chunk.subChunks[y] = NewSubChunk()
	return chunk.subChunks[y]
}

// SubChunkExists checks if the chunk has a sub chunk with the given Y value.
func (chunk *Chunk) SubChunkExists(y byte) bool {
	chunk.RLock()
	var _, ok = chunk.subChunks[y]
	chunk.RUnlock()
	return ok
}

// GetSubChunks returns all sub chunks in a Y => sub chunk map.
func (chunk *Chunk) GetSubChunks() map[byte]*SubChunk {
	return chunk.subChunks
}

// SetHeightMapAt sets the height map at the given column to the given value.
func (chunk *Chunk) SetHeightMapAt(x, z int, value int16) {
	chunk.HeightMap[chunk.GetHeightMapIndex(x, z)] = value
}

// GetHeightMapAt returns the height map value at the given column.
func (chunk *Chunk) GetHeightMapAt(x, z int) int16 {
	return chunk.HeightMap[chunk.GetHeightMapIndex(x, z)]
}

// RecalculateHeightMap recalculates the height map.
func (chunk *Chunk) RecalculateHeightMap() {
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			chunk.SetHeightMapAt(x, z, chunk.GetHighestBlockY(x, z) + 1)
		}
	}
}

// GetHighestSubChunk returns the highest non-empty sub chunk index
func (chunk *Chunk) GetHighestSubChunkIndex() int {
	chunk.RLock()
	defer chunk.RUnlock()
	for y := 15; y >= 0; y-- {
		if _, ok := chunk.subChunks[byte(y)]; !ok {
			continue
		}
		if chunk.subChunks[byte(y)].IsAllAir() {
			continue
		}
		return y
	}
	return -1
}

// GetHighestSubChunk returns the highest non-empty sub chunk in the chunk.
func (chunk *Chunk) GetHighestSubChunk() *SubChunk {
	return chunk.subChunks[byte(chunk.GetHighestSubChunkIndex())]
}

// GetHighestBlockId returns the highest block ID at the given column.
func (chunk *Chunk) GetHighestBlockId(x, z int) byte {
	return chunk.GetHighestSubChunk().GetHighestBlockId(x, z)
}

// GetHighestBlockData returns the highest block data at the given column.
func (chunk *Chunk) GetHighestBlockData(x, z int) byte {
	return chunk.GetHighestSubChunk().GetHighestBlockData(x, z)
}

// GetFilledSubChunks returns the count of filled sub chunks in the chunk.
func (chunk *Chunk) GetFilledSubChunks() byte {
	//chunk.PruneEmptySubChunks()
	return byte(len(chunk.subChunks))
}

// PruneEmptySubChunks prunes all empty sub chunks that are not covered by filled ones.
// It works from up to down and returns immediately if a non-empty chunk was found.
func (chunk *Chunk) PruneEmptySubChunks() {
	chunk.Lock()
	for y := 15; y >= 0; y-- {
		if _, ok := chunk.subChunks[byte(y)]; !ok {
			continue
		}
		if !chunk.subChunks[byte(y)].IsAllAir() {
			break
		}
		delete(chunk.subChunks, byte(y))
	}
	chunk.Unlock()
}

func (chunk *Chunk) GetHighestBlockY(x, z int) int16 {
	var index = chunk.GetHighestSubChunkIndex()
	if index == -1 {
		return -1
	}
	var y int16
	for y = int16(index); y >= 0; y-- {
		var height = chunk.GetSubChunk(byte(y)).GetHighestBlockY(x, z) | (y << 4)
		if height != -1{
			return height
		}
	}
	return -1
}

// ToBinary converts the chunk to its binary representation, used for network sending.
func (chunk *Chunk) ToBinary() []byte {
	var stream = binutils.NewStream()
	var subChunkCount = chunk.GetFilledSubChunks()
	stream.PutByte(subChunkCount)
	//chunk.RLock()
	for i := byte(0); i < subChunkCount; i++ {
		if _, ok := chunk.subChunks[i]; !ok {
			stream.PutBytes(make([]byte, 4096+2048+1))
		} else {
			stream.PutBytes(chunk.subChunks[i].ToBinary())
		}
	}
	//chunk.RUnlock()
	for i := 255; i >= 0; i-- {
		stream.PutLittleShort(chunk.HeightMap[i])
	}
	for _, biome := range chunk.Biomes {
		stream.PutByte(byte(biome))
	}
	stream.PutByte(0)
	return stream.GetBuffer()
}

// GetBlockNBTIndex returns the block NBT index of the given X, Y and Z.
func GetBlockNBTIndex(x, y, z int) int {
	return ((y & 256) << 8) | ((x & 15) << 4) | (z & 15)
}
