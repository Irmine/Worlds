package worlds

import "github.com/irmine/worlds/chunks"

// Loader is a struct used to load a range of chunks from a dimension.
type Loader struct {
	Dimension *Dimension
	ChunkX    int32
	ChunkZ    int32
	loadedChunks map[int]*chunks.Chunk
	LoadFunction func(*chunks.Chunk)
}

// NewLoader returns a new loader on the given dimension with the given chunk X and Z.
func NewLoader(dimension *Dimension, x, z int32) *Loader {
	return &Loader{dimension, x, z, make(map[int]*chunks.Chunk), func(chunk *chunks.Chunk){}}
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

// HasChunkInUse checks if the loader has a chunk with the given chunk X and Z in use.
func (loader *Loader) HasChunkInUse(chunkX, chunkZ int32) bool {
	var _, ok = loader.loadedChunks[GetChunkIndex(chunkX, chunkZ)]
	return ok
}

// setChunkInUse sets the given chunk in use.
func (loader *Loader) setChunkInUse(chunkX, chunkZ int32, chunk *chunks.Chunk) {
	loader.loadedChunks[GetChunkIndex(chunkX, chunkZ)] = chunk
}

// Request requests all chunks within the given view distance from the current position.
// All chunks loaded will run the load function of this loader.
func (loader *Loader) Request(distance int32) {
	var f = func(chunk *chunks.Chunk) {
		loader.setChunkInUse(chunk.GetX(), chunk.GetZ(), chunk)
		loader.LoadFunction(chunk)
	}
	for x := -distance + loader.ChunkX; x <= distance+loader.ChunkX; x++ {
		for z := -distance + loader.ChunkZ; z <= distance+loader.ChunkZ; z++ {

			var xRel = x - loader.ChunkX
			var zRel = z - loader.ChunkZ
			if xRel*xRel+zRel*zRel <= distance*distance {
				if loader.HasChunkInUse(x, z) {
					continue
				}

				if !loader.Dimension.chunkProvider.IsChunkLoaded(x, z) {
					loader.Dimension.chunkProvider.LoadChunk(x, z, f)
				} else {
					chunk, _ := loader.Dimension.GetChunk(x, z)
					f(chunk)
				}
			}
		}
	}
}
