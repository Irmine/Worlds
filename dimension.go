package worlds

import (
	"errors"
	"github.com/golang/geo/r3"
	"github.com/google/uuid"
	"github.com/irmine/worlds/blocks"
	"github.com/irmine/worlds/chunks"
	"github.com/irmine/worlds/generation"
	"github.com/irmine/worlds/providers"
	"github.com/irmine/worlds/utils"
	"math"
	"os"
	"sync"
)

const (
	OverworldId DimensionId = iota
	NetherId
	EndId
)

// DimensionId is an ID of a dimension.
type DimensionId byte

// Dimension is a struct which holds helper functions for chunks.
type Dimension struct {
	name  string
	level *Level
	id    DimensionId

	chunkProvider providers.Provider
	blockManager  blocks.Manager

	mutex    sync.RWMutex
	entities map[uint64]chunks.ChunkEntity
	viewers  map[uuid.UUID]chunks.Viewer

	blockUpdates map[int64]r3.Vector
}

// EntityRuntimeId is an ever increasing unsigned int64.
// Every entity placed in the world increments the runtime ID.
var EntityRuntimeId uint64

// UnloadedChunk gets returned if a block is attempted to be retrieved from an unloaded chunk.
var UnloadedChunk = errors.New("chunk is not loaded")

// UnavailableEntity gets returned if an entity is attempted to be retrieved but is not available in the dimension.
var UnavailableEntity = errors.New("dimension does not have entity with runtime ID available")

// NewDimension returns a new dimension with the given name, levelName, dimension ID and server path.
// Dimension data will be written in the `serverPath/worlds/levelName/name` path.
func NewDimension(name string, level *Level, id DimensionId) *Dimension {
	var path = level.serverPath + "worlds/" + level.GetName() + "/" + name + "/region/"
	os.MkdirAll(path, 0700)

	var dimension = &Dimension{name, level, id, nil, nil, sync.RWMutex{}, make(map[uint64]chunks.ChunkEntity), make(map[uuid.UUID]chunks.Viewer), make(map[int64]r3.Vector)}

	return dimension
}

// GetDimensionId returns the dimension ID of the dimension.
func (dimension *Dimension) GetDimensionId() DimensionId {
	return dimension.id
}

// GetName returns the name of the dimension.
func (dimension *Dimension) GetName() string {
	return dimension.name
}

// GetLevel returns the level of the dimension.
func (dimension *Dimension) GetLevel() *Level {
	return dimension.level
}

// Close closes the dimension and saves it.
// If async is true, closes the dimension asynchronously.
func (dimension *Dimension) Close(async bool) {
	dimension.chunkProvider.Close(async)
}

// Save saves the dimension.
func (dimension *Dimension) Save() {
	dimension.chunkProvider.Save()
}

// GetEntities returns all loaded entities in this dimension in a runtime ID => entity map.
func (dimension *Dimension) GetEntities() map[uint64]chunks.ChunkEntity {
	return dimension.entities
}

// GetViewers returns all entities considered as viewers in the dimension.
func (dimension *Dimension) GetViewers() map[uuid.UUID]chunks.Viewer {
	return dimension.viewers
}

// AddViewer adds a viewer to the dimension.
func (dimension *Dimension) AddViewer(viewer chunks.Viewer, position r3.Vector) {
	x, z := int32(math.Floor(position.X))>>4, int32(math.Floor(position.Z))>>4
	dimension.LoadChunk(x, z, func(chunk *chunks.Chunk) {
		dimension.mutex.Lock()
		dimension.viewers[viewer.GetUUID()] = viewer
		dimension.mutex.Unlock()
		chunk.AddViewer(viewer)
	})
}

// RemoveViewer removes a viewer from the dimension.
func (dimension *Dimension) RemoveViewer(uuid uuid.UUID) {
	dimension.mutex.Lock()
	delete(dimension.viewers, uuid)
	dimension.mutex.Unlock()
}

// GetViewer returns a viewer of a dimension by its UUID.
// A bool gets returned indicating whether the viewer was found or not.
func (dimension *Dimension) GetViewer(uuid uuid.UUID) (chunks.Viewer, bool) {
	dimension.mutex.RLock()
	viewer, ok := dimension.viewers[uuid]
	dimension.mutex.RUnlock()
	return viewer, ok
}

// AddEntity adds a new entity at the given position in the dimension.
func (dimension *Dimension) AddEntity(entity chunks.ChunkEntity, position r3.Vector) {
	var x, z = int32(math.Floor(position.X)) >> 4, int32(math.Floor(position.Z)) >> 4
	dimension.LoadChunk(x, z, func(chunk *chunks.Chunk) {
		EntityRuntimeId++
		entity.SetRuntimeId(EntityRuntimeId)
		entity.SetDimension(dimension)
		entity.SetPosition(position)
		entity.SpawnToAll()

		chunk.AddEntity(entity)
		dimension.mutex.Lock()
		dimension.entities[EntityRuntimeId] = entity
		dimension.mutex.Unlock()
	})
}

// RemoveEntity removes an entity in the dimension with the given runtime ID.
// The removed entity also gets closed if not yet done.
func (dimension *Dimension) RemoveEntity(runtimeId uint64) {
	dimension.mutex.Lock()
	if entity, ok := dimension.entities[runtimeId]; ok {
		if !entity.IsClosed() {
			entity.Close()
		}
		var x, z = int32(math.Floor(entity.GetPosition().X)), int32(math.Floor(entity.GetPosition().Z))
		if chunk, ok := dimension.GetChunk(x, z); ok {
			chunk.RemoveEntity(runtimeId)
		}
		delete(dimension.entities, runtimeId)
	}
	dimension.mutex.Unlock()
}

// GetEntity returns an entity in the dimension by its runtime ID.
// Returns UnavailableEntity error if no entity with that runtime ID was available in the dimension.
func (dimension *Dimension) GetEntity(runtimeId uint64) (chunks.ChunkEntity, error) {
	dimension.mutex.RLock()
	defer dimension.mutex.RUnlock()
	if entity, ok := dimension.entities[runtimeId]; ok {
		return entity, nil
	}
	return nil, UnavailableEntity
}

// HasEntity checks if the dimension has an entity available with the given runtime ID.
func (dimension *Dimension) HasEntity(runtimeId uint64) bool {
	dimension.mutex.RLock()
	var _, ok = dimension.entities[runtimeId]
	dimension.mutex.RUnlock()
	return ok
}

// IsChunkLoaded checks if a chunk at the given chunk X and Z is loaded.
func (dimension *Dimension) IsChunkLoaded(x, z int32) bool {
	return dimension.chunkProvider.IsChunkLoaded(x, z)
}

// UnloadChunk unloads a chunk at the given chunk X and Z.
func (dimension *Dimension) UnloadChunk(x, z int32) {
	dimension.chunkProvider.UnloadChunk(x, z)
}

// LoadChunk submits a request with the given chunk X and Z to get loaded.
// The function given gets run as soon as the chunk gets loaded.
func (dimension *Dimension) LoadChunk(x, z int32, function func(chunk *chunks.Chunk)) {
	dimension.chunkProvider.LoadChunk(x, z, function)
}

// SetChunk sets a new chunk at the given chunk X and Z.
func (dimension *Dimension) SetChunk(x, z int32, chunk *chunks.Chunk) {
	dimension.chunkProvider.SetChunk(x, z, chunk)
}

// GetChunk returns a chunk in the dimension at the given chunk X and Z.
func (dimension *Dimension) GetChunk(x, z int32) (*chunks.Chunk, bool) {
	return dimension.chunkProvider.GetChunk(x, z)
}

// SetGenerator sets the generator of the dimension.
func (dimension *Dimension) SetGenerator(generator generation.Generator) {
	dimension.chunkProvider.SetGenerator(generator)
}

// GetGenerator returns the generator of the dimension.
func (dimension *Dimension) GetGenerator() generation.Generator {
	return dimension.chunkProvider.GetGenerator()
}

// GetChunkProvider returns the chunk provider of the dimension.
func (dimension *Dimension) GetChunkProvider() providers.Provider {
	return dimension.chunkProvider
}

// SetChunkProvider sets the chunk provider of the dimension.
func (dimension *Dimension) SetChunkProvider(provider providers.Provider) {
	dimension.chunkProvider = provider
}

// GetBlockAt returns a block in the dimension at the given vector.
// GetBlockAt returns an error when the chunk of the block was not loaded, and an error if a block with the given ID wasn't registered.
func (dimension *Dimension) GetBlockAt(vector r3.Vector) (blocks.Block, error) {
	var x, y, z = int(math.Floor(vector.X)), int(math.Floor(vector.Y)), int(math.Floor(vector.Z))
	var chunk, ok = dimension.GetChunk(int32(x>>4), int32(z>>4))
	if !ok {
		return nil, UnloadedChunk
	}
	var id, meta = chunk.GetBlockId(x&15, y, z&15), chunk.GetBlockData(x&15, y, z&15)
	var block, err = dimension.blockManager.Get(id, meta)
	if err != nil {
		if nbt, ok := chunk.GetBlockNBTAt(x&15, y, z&15); ok {
			block.SetNBT(nbt)
		}
	}
	return block, err
}

// SetBlockAt sets a block at the given vector.
// If the chunk at that position was not yet loaded, it loads it and places the block.
func (dimension *Dimension) SetBlockAt(vector r3.Vector, block blocks.Block) {
	var x, y, z = int(math.Floor(vector.X)), int(math.Floor(vector.Y)), int(math.Floor(vector.Z))
	dimension.LoadChunk(int32(x>>4), int32(z>>4), func(chunk *chunks.Chunk) {
		chunk.SetBlockId(x&15, y, z&15, block.GetId())
		chunk.SetBlockData(x&15, y, z&15, block.GetData())
		chunk.SetBlockNBTAt(x&15, y, z&15, block.GetNBT())
		dimension.SetBlockForUpdate(vector)
	})
}

// HasBlockUpdates sets a block at a certain position to be updated on the next tick
func (dimension *Dimension) HasBlockUpdates() bool {
	return len(dimension.blockUpdates) > 0
}

// SetBlockUpdate sets a block at a certain position to be updated on the next tick
func (dimension *Dimension) SetBlockForUpdate(vector r3.Vector) {
	var x, y, z = int64(math.Floor(vector.X)), int64(math.Floor(vector.Y)), int64(math.Floor(vector.Z))
	dimension.blockUpdates[(x << 12) | (z << 8) | y] = vector
}

// SetBlockUpdate removes a block at a certain position from being updated
func (dimension *Dimension) UnsetBlockFromUpdate(vector r3.Vector) {
	var x, y, z = int64(math.Floor(vector.X)), int64(math.Floor(vector.Y)), int64(math.Floor(vector.Z))
	var index = (x << 12) | (z << 8) | y
	if _, ok := dimension.blockUpdates[index]; ok {
		delete(dimension.blockUpdates, index)
	}
}

// getUpdatedBlocks returns all the blocks and runtime ids that need to be updated
// the first return value is all the positions of the blocks that need to be changed
// the second return value is all the runtime ids for the blocks that need to be changed
func (dimension *Dimension) getUpdatedBlocks() ([]blocks.Position, []uint32) {
	var runtimeIds []uint32
	var position []blocks.Position
	for index, vector := range dimension.blockUpdates {
		var x, y, z = int(math.Floor(vector.X)), int(math.Floor(vector.Y)), int(math.Floor(vector.Z))
		dimension.LoadChunk(int32(x>>4), int32(z>>4), func(chunk *chunks.Chunk) {
			blockId := int(chunk.GetBlockId(x&15, y, z&15))
			blockData := int(chunk.GetBlockData(x&15, y, z&15))

			if runtimeId, ok := blocks.GetRuntimeId(blockId, blockData); ok {
				runtimeIds = append(runtimeIds, runtimeId)
				position = append(position, utils.VectorToPosition(vector))
			}
			delete(dimension.blockUpdates, index)
		})
	}
	return position, runtimeIds
}

// ProcessBlockUpdates processes all the block update requests
func (dimension *Dimension) ProcessBlockUpdates() {
	var positions, runtimeIds = dimension.getUpdatedBlocks()
	for index := range positions {
		for _, viewer := range dimension.GetViewers() {
			viewer.SendUpdateBlock(positions[index], runtimeIds[index], 0)
		}
	}
}

// Tick ticks the entire dimension, such as entities.
func (dimension *Dimension) Tick() {
	if dimension.HasBlockUpdates() {
		dimension.ProcessBlockUpdates()
	}
	for runtimeId, entity := range dimension.entities {
		if entity.IsClosed() {
			dimension.RemoveEntity(runtimeId)
		} else {
			entity.Tick()
		}
	}
}
