package providers

import (
	"github.com/irmine/worlds/chunks"
	"github.com/irmine/worlds/generation"
	"sync"
)

// Provider is the interface used to manage chunks and generators.
type Provider interface {
	Save()
	Close(bool)
	LoadChunk(int32, int32, func(*chunks.Chunk))
	IsChunkLoaded(int32, int32) bool
	UnloadChunk(int32, int32)
	SetChunk(int32, int32, *chunks.Chunk)
	GetChunk(int32, int32) (*chunks.Chunk, bool)
	SetGenerator(generation.Generator)
	GetGenerator() generation.Generator
	GenerateChunk(int32, int32)
}

// ChunkProvider implements the Provider interface, implementing basic functionality of a chunk provider.
type ChunkProvider struct {
	generator generation.Generator
	requests  chan ChunkRequest

	mutex  sync.RWMutex
	chunks map[int]*chunks.Chunk
}

// ChunkRequest is a struct used to request a chunk and execute a function once loaded.
type ChunkRequest struct {
	function func(*chunks.Chunk)
	x        int32
	z        int32
}

// New returns a new chunk provider.
func new() *ChunkProvider {
	return &ChunkProvider{requests: make(chan ChunkRequest, 4096), chunks: make(map[int]*chunks.Chunk)}
}

// LoadChunk loads the chunk at the given chunk X and Z.
// The function provided will run with the loaded chunk once done.
// The function gets ran immediately if the chunk is already loaded.
func (provider *ChunkProvider) LoadChunk(x, z int32, function func(*chunks.Chunk)) {
	if chunk, ok := provider.GetChunk(x, z); ok {
		function(chunk)
		return
	}
	provider.requests <- ChunkRequest{function, x, z}
}

// IsChunkLoaded checks if a chunk is loaded at the given chunk X and Z.
func (provider *ChunkProvider) IsChunkLoaded(x, z int32) bool {
	provider.mutex.RLock()
	var _, ok = provider.chunks[provider.GetChunkIndex(x, z)]
	provider.mutex.RUnlock()
	return ok
}

// UnloadChunk unloads a chunk with the given chunk X and Z if loaded.
func (provider *ChunkProvider) UnloadChunk(x, z int32) {
	provider.mutex.Lock()
	delete(provider.chunks, provider.GetChunkIndex(x, z))
	provider.mutex.Unlock()
}

// SetChunk sets a chunk at the given chunk X and Z.
func (provider *ChunkProvider) SetChunk(x, z int32, chunk *chunks.Chunk) {
	provider.mutex.Lock()
	provider.chunks[provider.GetChunkIndex(x, z)] = chunk
	provider.mutex.Unlock()
}

// GetChunk returns the chunk at the given chunk X and Z.
// Returns false if no loaded chunk was found at that position.
func (provider *ChunkProvider) GetChunk(x, z int32) (*chunks.Chunk, bool) {
	provider.mutex.RLock()
	var chunk, ok = provider.chunks[provider.GetChunkIndex(x, z)]
	provider.mutex.RUnlock()
	return chunk, ok
}

// SetGenerator sets the generator of the provider.
func (provider *ChunkProvider) SetGenerator(generator generation.Generator) {
	provider.generator = generator
}

// GetGenerator returns the generator of the provider.
func (provider *ChunkProvider) GetGenerator() generation.Generator {
	return provider.generator
}

// completeRequest completes the given request, executing its function.
func (provider *ChunkProvider) completeRequest(request ChunkRequest) {
	var chunk, ok = provider.GetChunk(request.x, request.z)
	if ok {
		request.function(chunk)
	}
}

// GenerateChunk generates a new chunk at the given chunk X and Z.
func (provider *ChunkProvider) GenerateChunk(x, z int32) {
	var chunk = provider.generator.GenerateNewChunk(x, z)
	provider.SetChunk(x, z, chunk)
}

// GetChunkIndex returns the chunk index of the given chunk X and Z.
func (provider *ChunkProvider) GetChunkIndex(x, z int32) int {
	return int(((int64(x) & 0xffffffff) << 32) | (int64(z) & 0xffffffff))
}
