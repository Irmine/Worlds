package worlds

import (
	"github.com/irmine/worlds/chunks"
	"sync"
)

// Loader is a struct used to load a range of chunks from a dimension.
type Loader struct {
	Dimension *Dimension
	ChunkX    int32
	ChunkZ    int32

	// UnloadFunction gets called for every chunk that gets unloaded in this loader.
	// Chunks that exceed the request range automatically get unloaded during request.
	UnloadFunction func(*chunks.Chunk)
	// LoadFunction gets called for every chunk loaded by this loader.
	LoadFunction func(*chunks.Chunk)

	mutex        sync.RWMutex
	loadedChunks map[int]*chunks.Chunk
}

// NewLoader returns a new loader on the given dimension with the given chunk X and Z.
func NewLoader(dimension *Dimension, x, z int32) *Loader {
	return &Loader{dimension, x, z, func(chunk *chunks.Chunk) {}, func(chunk *chunks.Chunk) {}, sync.RWMutex{}, make(map[int]*chunks.Chunk)}
}

// Move moves the loader to the given chunk X and Z.
func (loader *Loader) Move(chunkX, chunkZ int32) {
	loader.ChunkX = chunkX
	loader.ChunkZ = chunkZ
}

// Warp warps the loader to the given dimension and moves it to the given chunk X and Z.
func (loader *Loader) Warp(dimension *Dimension, chunkX, chunkZ int32) {
	loader.Dimension = dimension
	loader.Move(chunkX, chunkZ)
}

// GetLoadedChunkCount returns the count of the loaded chunks.
func (loader *Loader) GetLoadedChunkCount() int {
	return len(loader.loadedChunks)
}

// HasChunkInUse checks if the loader has a chunk with the given chunk X and Z in use.
func (loader *Loader) HasChunkInUse(chunkX, chunkZ int32) bool {
	loader.mutex.RLock()
	var _, ok = loader.loadedChunks[GetChunkIndex(chunkX, chunkZ)]
	loader.mutex.RUnlock()
	return ok
}

// setChunkInUse sets the given chunk in use.
func (loader *Loader) setChunkInUse(chunkX, chunkZ int32, chunk *chunks.Chunk) {
	loader.mutex.Lock()
	loader.loadedChunks[GetChunkIndex(chunkX, chunkZ)] = chunk
	loader.mutex.Unlock()
}

// Request requests all chunks within the given view distance from the current position.
// All chunks loaded will run the load function of this loader.
// Request will also unload any unused chunks beyond the distance specified.
func (loader *Loader) Request(distance int32, maximumChunks int) {
	var f = func(chunk *chunks.Chunk) {
		loader.setChunkInUse(chunk.X, chunk.Z, chunk)
		loader.LoadFunction(chunk)
	}
	i := 0
	for x := -distance + loader.ChunkX; x <= distance+loader.ChunkX; x++ {
		for z := -distance + loader.ChunkZ; z <= distance+loader.ChunkZ; z++ {
			if i == maximumChunks {
				break
			}
			var xRel = x - loader.ChunkX
			var zRel = z - loader.ChunkZ
			if xRel*xRel+zRel*zRel <= distance*distance {
				if loader.HasChunkInUse(x, z) {
					continue
				}

				i++
				loader.Dimension.chunkProvider.LoadChunk(x, z, f)
			}
		}
	}
	loader.unloadUnused(distance)
}

// unloadUnused unloads all unused chunks beyond the given distance.
func (loader *Loader) unloadUnused(distance int32) {
	var rs = distance * distance
	loader.mutex.Lock()
	for index, chunk := range loader.loadedChunks {
		xDist := loader.ChunkX - chunk.X
		zDist := loader.ChunkZ - chunk.Z

		if xDist*xDist+zDist*zDist > rs {
			delete(loader.loadedChunks, index)
			loader.UnloadFunction(chunk)
		}
	}
	loader.mutex.Unlock()
}
