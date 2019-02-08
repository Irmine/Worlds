package worlds

import (
	"github.com/irmine/worlds/chunks"
	"math"
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
	LoadFunction            func(*chunks.Chunk)
	PublisherUpdateFunction func()

	mutex sync.RWMutex

	loadedChunks     map[int]*chunks.Chunk
	loadChunkQueue   map[int]bool
	unloadChunkQueue map[int]bool
}

// NewLoader returns a new loader on the given dimension with the given chunk X and Z.
func NewLoader(dimension *Dimension, x, z int32) *Loader {
	return &Loader{dimension, x, z, func(chunk *chunks.Chunk) {}, func(chunk *chunks.Chunk) {}, func(){}, sync.RWMutex{}, make(map[int]*chunks.Chunk), make(map[int]bool), make(map[int]bool)}
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

// GetLoadedChunkCount returns loaded chunks.
func (loader *Loader) GetLoadedChunkCount() int {
	return len(loader.loadedChunks)
}

// GetLoadedChunkCount returns the count of the loaded chunks.
func (loader *Loader) GetLoadedChunks() map[int]*chunks.Chunk {
	return loader.loadedChunks
}

// HasChunkInUse checks if the loader has a chunk with the given chunk X and Z in use.
func (loader *Loader) HasChunkInUse(chunkX, chunkZ int32) bool {
	loader.mutex.RLock()
	var _, ok = loader.loadedChunks[loader.GetChunkHash(chunkX, chunkZ)]
	loader.mutex.RUnlock()
	return ok
}

// setChunkInUse sets the given chunk in use.
func (loader *Loader) setChunkInUse(chunkX, chunkZ int32, chunk *chunks.Chunk) {
	loader.mutex.Lock()
	loader.loadedChunks[loader.GetChunkHash(chunkX, chunkZ)] = chunk
	loader.mutex.Unlock()
}

// unsetChunkFromUse removes a chunk from use
func (loader *Loader) unsetChunkFromUse(chunkX, chunkZ int32) {
	loader.mutex.Lock()
	delete(loader.loadedChunks, loader.GetChunkHash(chunkX, chunkZ))
	loader.mutex.Unlock()
}

// Returns chunk coordinates as a identifiable hash
func (loader *Loader) GetChunkHash(x, z int32) int {
	return loader.Dimension.chunkProvider.GetChunkIndex(x, z)
}

// Returns chunk coordinates from an identifiable hash
func (loader *Loader) GetChunkXZ(hash int) (int, int) {
	return loader.Dimension.chunkProvider.GetChunkXZ(hash)
}

// Puts a chunk in the load queue to be processed
func (loader *Loader) LoadChunk(x, z int32) {
	var Hash = loader.GetChunkHash(x, z)
	if !loader.HasChunkInUse(x, z) {
		loader.loadChunkQueue[Hash] = true
	}
	delete(loader.unloadChunkQueue, Hash)
}

// Puts a chunk in the unload queue to be processed
func (loader *Loader) UnloadChunk(x, z int32) {
	var Hash = loader.GetChunkHash(x, z)
	if loader.HasChunkInUse(x, z) {
		loader.unloadChunkQueue[Hash] = true
	}
	delete(loader.loadChunkQueue, Hash)
}

// SortChunks determines what chunks will be put in the chunk
// load queue and the unload queue based on the viewDistance given.
func (loader *Loader) SortChunks(viewDistance int32) {
	var ChunkX= loader.ChunkX
	var ChunkZ= loader.ChunkZ

	var RadiusSquared= int32(math.Pow(float64(viewDistance), 2))

	for index := range loader.loadedChunks{
		loader.unloadChunkQueue[index] = true
	}
	loader.loadChunkQueue = make(map[int]bool)

	for x := -viewDistance; x <= viewDistance; x++ {
		for z := -viewDistance; z <= viewDistance; z++ {
			if (int32(math.Pow(float64(x), 2)) + int32(math.Pow(float64(z), 2))) > RadiusSquared {
				continue
			}
			loader.LoadChunk(ChunkX+x, ChunkZ+z)
		}
	}
}

// ProcessLoadQueue processed all the chunk load queue with
// and optional chunks per tick limit
func (loader *Loader) ProcessLoadQueue(perTick int) {
	var f = func(chunk *chunks.Chunk) {
		loader.setChunkInUse(chunk.X, chunk.Z, chunk)
		loader.LoadFunction(chunk)
	}
	var count = 1
	for index := range loader.loadChunkQueue {
		if count >= perTick {
			break
		}
		var x, z= loader.GetChunkXZ(index)
		loader.Dimension.chunkProvider.LoadChunk(int32(x), int32(z), f)
		delete(loader.loadChunkQueue, index)
		count++
	}
}

// ProcessLoadQueue processed all the chunk unload queue with
// and optional chunks per tick limit
func (loader *Loader) ProcessUnloadQueue(perTick int) {
	loader.mutex.Lock()
	var count = 1
	for index := range loader.unloadChunkQueue {
		if count >= perTick {
			break
		}
		var x, z= loader.GetChunkXZ(index)
		var chunk, ok= loader.Dimension.chunkProvider.GetChunk(int32(x), int32(z))
		if ok {
			loader.UnloadFunction(chunk)
		}
		if _, ok := loader.loadedChunks[index]; ok {
			delete(loader.loadedChunks, index)
		}
		if _, ok := loader.loadChunkQueue[index]; ok {
			delete(loader.loadChunkQueue, index)
		}
		count++
	}
	loader.mutex.Unlock()
}

func (loader *Loader) Request(distance int32, perTick int) {
	loader.SortChunks(distance)
	if len(loader.loadChunkQueue) > 0 {
		loader.PublisherUpdateFunction()
		loader.ProcessLoadQueue(perTick)
	}
	if len(loader.unloadChunkQueue) > 0 {
		loader.ProcessUnloadQueue(perTick)
	}
}