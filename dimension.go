package worlds

import (
	"os"
	"github.com/irmine/worlds/providers"
	"github.com/irmine/worlds/chunks"
	"github.com/irmine/worlds/generation"
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
	name          string
	levelName	  string
	serverPath    string
	id            DimensionId
	chunkProvider providers.Provider
}

// NewDimension returns a new dimension with the given name, levelName, dimension ID and server path.
// Dimension data will be written in the `serverPath/worlds/levelName/name` path.
func NewDimension(name string, levelName string, id DimensionId, serverPath string) *Dimension {
	var path = serverPath + "worlds/" + levelName + "/" + name + "/region/"
	os.MkdirAll(path, 0700)

	var dimension = &Dimension{name, levelName, serverPath, id, nil}

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

// Close closes the dimension and saves it.
// If async is true, closes the dimension asynchronously.
func (dimension *Dimension) Close(async bool) {
	dimension.chunkProvider.Close(async)
}

// Save saves the dimension.
func (dimension *Dimension) Save() {
	dimension.chunkProvider.Save()
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

// TickDimension ticks the entire dimension.
func (dimension *Dimension) TickDimension() {

}
