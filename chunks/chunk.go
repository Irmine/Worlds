package chunks

import (
	"errors"
	"github.com/irmine/binutils"
	"github.com/irmine/gomine/utils"
	"github.com/irmine/gonbt"
	"sync"
)

// Chunk is a segment of the world, holding blocks, block data, biomes, etc.
type Chunk struct {
	X, Z             int32
	subChunks        map[byte]*SubChunk
	LightPopulated   bool
	TerrainPopulated bool
	blockNBT         map[int]*gonbt.Compound
	entities         map[uint64]ChunkEntity
	Biomes           []byte
	HeightMap        []int16
	viewers          map[utils.UUID]Viewer
	InhabitedTime    int64
	LastUpdate       int64
	mutex            sync.RWMutex
}

// New returns a new chunk with the given X and Z.
func New(x, z int32) *Chunk {
	return &Chunk{
		x,
		z,
		make(map[byte]*SubChunk),
		true,
		true,
		make(map[int]*gonbt.Compound),
		make(map[uint64]ChunkEntity),
		make([]byte, 256),
		make([]int16, 256),
		make(map[utils.UUID]Viewer),
		0,
		0,
		sync.RWMutex{},
	}
}

// GetViewers returns all viewers of the chunk.
// Viewers are all players that have the chunk within their view distance.
func (chunk *Chunk) GetViewers() map[utils.UUID]Viewer {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return chunk.viewers
}

// AddViewer adds a viewer of the chunk.
func (chunk *Chunk) AddViewer(player Viewer) {
	chunk.mutex.Lock()
	chunk.viewers[player.GetUUID()] = player
	chunk.mutex.Unlock()
}

// RemoveViewer removes a viewer from the chunk.
func (chunk *Chunk) RemoveViewer(player Viewer) {
	chunk.mutex.Lock()
	delete(chunk.viewers, player.GetUUID())
	chunk.mutex.Unlock()
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
	chunk.mutex.Lock()
	chunk.entities[entity.GetRuntimeId()] = entity
	chunk.mutex.Unlock()
	return nil
}

// RemoveEntity removes an entity with the given runtimeId from the chunk.
func (chunk *Chunk) RemoveEntity(runtimeId uint64) {
	chunk.mutex.Lock()
	delete(chunk.entities, runtimeId)
	chunk.mutex.Unlock()
}

// GetEntities returns all entities of the chunk.
func (chunk *Chunk) GetEntities() map[uint64]ChunkEntity {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
	return chunk.entities
}

// SetBlockNBTAt sets the given compound at the given position.
func (chunk *Chunk) SetBlockNBTAt(x, y, z int, nbt *gonbt.Compound) {
	chunk.mutex.Lock()
	if nbt == nil {
		delete(chunk.blockNBT, GetBlockNBTIndex(x, y, z))
	} else {
		chunk.blockNBT[GetBlockNBTIndex(x, y, z)] = nbt
	}
	chunk.mutex.Unlock()
}

// RemoveBlockNBTAt removes the block NBT at the given position.
func (chunk *Chunk) RemoveBlockNBTAt(x, y, z int) {
	chunk.mutex.Lock()
	delete(chunk.blockNBT, GetBlockNBTIndex(x, y, z))
	chunk.mutex.Unlock()
}

// BlockNBTExistsAt checks if any block NBT exists at the given position.
func (chunk *Chunk) BlockNBTExistsAt(x, y, z int) bool {
	chunk.mutex.RLock()
	var _, ok = chunk.blockNBT[GetBlockNBTIndex(x, y, z)]
	chunk.mutex.RUnlock()
	return ok
}

// GetBlockNBTAt returns the block NBT at the given position.
// Returns a bool if any block NBT was found at that position
func (chunk *Chunk) GetBlockNBTAt(x, y, z int) (*gonbt.Compound, bool) {
	chunk.mutex.RLock()
	var c, ok = chunk.blockNBT[GetBlockNBTIndex(x, y, z)]
	chunk.mutex.RUnlock()
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
	chunk.mutex.Lock()
	chunk.subChunks[y] = subChunk
	chunk.mutex.Unlock()
}

// GetSubChunk returns a SubChunk on a given height index in this chunk.
func (chunk *Chunk) GetSubChunk(y byte) *SubChunk {
	if chunk.SubChunkExists(y) {
		chunk.mutex.RLock()
		defer chunk.mutex.RUnlock()
		return chunk.subChunks[y]
	}
	var sub = NewSubChunk()
	chunk.mutex.Lock()
	chunk.subChunks[y] = sub
	chunk.mutex.Unlock()
	return sub
}

// SubChunkExists checks if the chunk has a sub chunk with the given Y value.
func (chunk *Chunk) SubChunkExists(y byte) bool {
	chunk.mutex.RLock()
	var _, ok = chunk.subChunks[y]
	chunk.mutex.RUnlock()
	return ok
}

// GetSubChunks returns all sub chunks in a Y => sub chunk map.
func (chunk *Chunk) GetSubChunks() map[byte]*SubChunk {
	chunk.mutex.RLock()
	defer chunk.mutex.RUnlock()
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
			chunk.SetHeightMapAt(x, z, chunk.GetHighestSubChunk().GetHighestBlockY(x, z)+1)
		}
	}
}

// GetHighestSubChunk returns the highest non-empty sub chunk in the chunk.
func (chunk *Chunk) GetHighestSubChunk() *SubChunk {
	chunk.mutex.RLock()
	for y := byte(15); y >= 0; y-- {
		if _, ok := chunk.subChunks[y]; !ok {
			continue
		}
		if chunk.subChunks[y].IsAllAir() {
			continue
		}
		return chunk.subChunks[y]
	}
	chunk.mutex.RUnlock()
	return nil
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
	chunk.PruneEmptySubChunks()
	return byte(len(chunk.subChunks))
}

// PruneEmptySubChunks prunes all empty sub chunks that are not covered by filled ones.
// It works from up to down and returns immediately if a non-empty chunk was found.
func (chunk *Chunk) PruneEmptySubChunks() {
	chunk.mutex.Lock()
	for y := byte(15); y >= 0; y-- {
		if _, ok := chunk.subChunks[y]; !ok {
			continue
		}
		if !chunk.subChunks[y].IsAllAir() {
			return
		}
		delete(chunk.subChunks, y)
	}
	chunk.mutex.Unlock()
}

// ToBinary converts the chunk to its binary representation, used for network sending.
func (chunk *Chunk) ToBinary() []byte {
	var stream = binutils.NewStream()
	var subChunkCount = chunk.GetFilledSubChunks()

	stream.PutByte(subChunkCount)

	chunk.mutex.RLock()
	for i := byte(0); i < subChunkCount; i++ {
		if _, ok := chunk.subChunks[i]; !ok {
			stream.PutBytes(make([]byte, 4096+2048+1))
		} else {
			stream.PutBytes(chunk.subChunks[i].ToBinary())
		}
	}
	chunk.mutex.RUnlock()

	for i := 255; i >= 0; i-- {
		stream.PutLittleShort(chunk.HeightMap[i])
	}

	for _, biome := range chunk.Biomes {
		stream.PutByte(byte(biome))
	}
	stream.PutByte(0)

	stream.PutVarInt(0)

	return stream.GetBuffer()
}

// GetBlockNBTIndex returns the block NBT index of the given X, Y and Z.
func GetBlockNBTIndex(x, y, z int) int {
	return ((y & 256) << 8) | ((x & 15) << 4) | (z & 15)
}
