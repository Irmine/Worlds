package chunks

import (
	"sync"
)

// Chunk is a segment of the world, holding blocks, block data, biomes, etc.
type Chunk struct {
	X, Z             int32
	subChunks        map[byte]*SubChunk
	LightPopulated   bool
	TerrainPopulated bool
	tiles            map[uint64]tiles.Tile
	entities         map[uint64]interfaces.IEntity
	Biomes           []byte
	HeightMap        [256]int16
	viewers          sync.Map
	InhabitedTime    int64
	LastUpdate       int64
}

// New returns a new chunk with the given X and Z.
func New(x, z int32) *Chunk {
	return &Chunk{
		x,
		z,
		make(map[byte]*SubChunk),
		true,
		true,
		make(map[uint64]tiles.Tile),
		make(map[uint64]interfaces.IEntity),
		make([]byte, 256),
		[256]int16{},
		sync.Map{},
		0,
		0,
	}
}

// GetViewers returns all viewers of the chunk.
// Viewers are all players that have the chunk within their view distance.
func (chunk *Chunk) GetViewers() map[uint64]interfaces.IPlayer {
	var players = map[uint64]interfaces.IPlayer{}

	chunk.viewers.Range(func(runtimeId, player interface{}) bool {
		players[runtimeId.(uint64)] = player.(interfaces.IPlayer)
		return true
	})
	return players
}

// AddViewer adds a viewer of the chunk.
func (chunk *Chunk) AddViewer(player interfaces.IPlayer) {
	chunk.viewers.Store(player.GetRuntimeId(), player)
}

// RemoveViewer removes a viewer from the chunk.
func (chunk *Chunk) RemoveViewer(player interfaces.IPlayer) {
	chunk.viewers.Delete(player.GetRuntimeId())
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
func (chunk *Chunk) AddEntity(entity interfaces.IEntity) bool {
	if entity.IsClosed() {
		panic("Cannot add closed entity to chunk")
	}
	chunk.entities[entity.GetRuntimeId()] = entity
	return true
}

// RemoveEntity removes an entity with the given runtimeId from the chunk.
func (chunk *Chunk) RemoveEntity(runtimeId uint64) {
	if k, ok := chunk.entities[runtimeId]; ok {
		delete(chunk.entities, runtimeId)
	}
}

// GetEntities returns all entities of the chunk.
func (chunk *Chunk) GetEntities() map[uint64]interfaces.IEntity {
	return chunk.entities
}

func (chunk *Chunk) AddTile(tile tiles.Tile) bool {
	if tile.IsClosed() {
		panic("Cannot add closed entity to chunk")
	}
	chunk.tiles[tile.GetId()] = tile
	return true
}

func (chunk *Chunk) RemoveTile(tile tiles.Tile) {
	if k, ok := chunk.entities[tile.GetId()]; ok {
		delete(chunk.entities, k.GetRuntimeId())
	}
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
	chunk.GetSubChunk(byte(y >> 4)).SetBlockId(x, y&15, z, blockId)
}

// GetBlockId returns the block ID of a block at the given position.
func (chunk *Chunk) GetBlockId(x, y, z int) byte {
	return chunk.GetSubChunk(byte(y >> 4)).GetBlockId(x, y&15, z)
}

// SetBlockData sets the block data of a block at the given position.
func (chunk *Chunk) SetBlockData(x, y, z int, data byte) {
	chunk.GetSubChunk(byte(y >> 4)).SetBlockData(x, y&15, z, data)
}

// GetBlockData returns the block data of a block at the given position.
func (chunk *Chunk) GetBlockData(x, y, z int) byte {
	return chunk.GetSubChunk(byte(y >> 4)).GetBlockData(x, y&15, z)
}

// SetBlockLight sets the block light on a position in this chunk.
func (chunk *Chunk) SetBlockLight(x, y, z int, level byte) {
	chunk.GetSubChunk(byte(y >> 4)).SetBlockLight(x, y&15, z, level)
}

// GetBlockLight returns the block light on a position in this chunk.
func (chunk *Chunk) GetBlockLight(x, y, z int) byte {
	return chunk.GetSubChunk(byte(y >> 4)).GetBlockLight(x, y&15, z)
}

// SetSkyLight sets the sky light on a position in this chunk.
func (chunk *Chunk) SetSkyLight(x, y, z int, level byte) {
	chunk.GetSubChunk(byte(y >> 4)).SetSkyLight(x, y&15, z, level)
}

// GetSkyLight returns the sky light on a position in this chunk.
func (chunk *Chunk) GetSkyLight(x, y, z int) byte {
	return chunk.GetSubChunk(byte(y >> 4)).GetSkyLight(x, y&15, z)
}

// SetSubChunk sets a SubChunk on a position in this chunk.
func (chunk *Chunk) SetSubChunk(y byte, subChunk *SubChunk) {
	chunk.subChunks[y] = subChunk
}

// GetSubChunk returns a SubChunk on a given height index in this chunk.
func (chunk *Chunk) GetSubChunk(y byte) *SubChunk {
	if _, ok := chunk.subChunks[y]; ok {
		return chunk.subChunks[y]
	}
	chunk.subChunks[y] = NewSubChunk()
	return chunk.subChunks[y]
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
			chunk.SetHeightMapAt(x, z, chunk.GetHighestSubChunk().GetHighestBlockY(x, z)+1)
		}
	}
}

// GetHighestSubChunk returns the highest non-empty sub chunk in the chunk.
func (chunk *Chunk) GetHighestSubChunk() *SubChunk {
	for y := byte(15); y >= 0; y-- {
		if _, ok := chunk.subChunks[y]; ! ok {
			continue
		}
		if chunk.subChunks[y].IsAllAir() {
			continue
		}
		return chunk.subChunks[y]
	}
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
func (chunk *Chunk) PruneEmptySubChunks() {
	for y := byte(15); y >= 0; y-- {
		if _, ok := chunk.subChunks[y]; ! ok {
			continue
		}
		if !chunk.subChunks[y].IsAllAir() {
			return
		}
		delete(chunk.subChunks, y)
	}
}

// ToBinary converts the chunk to its binary representation, used for network sending.
func (chunk *Chunk) ToBinary() []byte {
	var stream = utils.NewStream()
	var subChunkCount = chunk.GetFilledSubChunks()

	stream.PutByte(subChunkCount)
	for i := byte(0); i < subChunkCount; i++ {
		if _, ok := chunk.subChunks[i]; !ok {
			stream.PutBytes(make([]byte, 4096+2048+1))
		} else {
			stream.PutBytes(chunk.subChunks[i].ToBinary())
		}
	}

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
