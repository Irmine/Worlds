package providers

import (
	"sync"
	"os"
	"strconv"
	"github.com/irmine/worlds/io"
	"github.com/irmine/nbt"
)

// Anvil is a provider for the MCAnvil world format.
// It uses the `.mca` file extension for region files.
type Anvil struct {
	path    string
	regions sync.Map

	*ChunkProvider
}

// NewAnvil returns an anvil chunk provider writing and reading regions from the given path.
func NewAnvil(path string) *Anvil {
	var provider = &Anvil{path, sync.Map{}, new()}
	go provider.Process()

	return provider
}

// Process continuously processes chunk requests for chunks that were not yet loaded when requested.
func (provider *Anvil) Process() {
	for {
		var request = <-provider.requests
		if provider.IsChunkLoaded(request.x, request.z) {
			provider.completeRequest(request)
			continue
		}

		go func() {
			var regionX, regionZ = request.x>>5, request.z>>5
			if provider.IsRegionLoaded(regionX, regionZ) {
				provider.load(request, regionX, regionZ)
			} else {
				var path = provider.path + "r." + strconv.Itoa(int(regionX)) + "." + strconv.Itoa(int(regionZ)) + ".mca"
				var _, err = os.Stat(path)
				if err != nil {
					os.Create(path)
				}
				provider.OpenRegion(regionX, regionZ, path)
				provider.load(request, regionX, regionZ)
			}
		}()
	}
}

// load loads a chunk at the given region X and Z for the given request.
func (provider *Anvil) load(request ChunkRequest, regionX, regionZ int32) {
	var region, _ = provider.GetRegion(regionX, regionZ)
	if !region.HasChunkGenerated(request.x, request.z) {
		provider.GenerateChunk(request.x, request.z)
		provider.completeRequest(request)
		return
	}

	var compression, data = region.GetChunkData(request.x, request.z)

	var reader = nbt.NewNBTReader(data, false, nbt.BigEndian)
	var c = reader.ReadIntoCompound(int(compression))

	if c == nil {
		provider.GenerateChunk(request.x, request.z)
		provider.completeRequest(request)
		return
	}

	provider.SetChunk(request.x, request.z, io.GetAnvilChunkFromNBT(c))
	provider.completeRequest(request)
}

// IsRegionLoaded checks if a region with the given region X and Z is loaded.
func (provider *Anvil) IsRegionLoaded(regionX, regionZ int32) bool {
	var _, ok = provider.regions.Load(provider.GetChunkIndex(regionX, regionZ))
	return ok
}

// GetRegion returns a region with the given region X and Z, or nil if it is not loaded, and a bool indicating success.
func (provider *Anvil) GetRegion(regionX, regionZ int32) (*io.Region, bool) {
	var region, ok = provider.regions.Load(provider.GetChunkIndex(regionX, regionZ))
	return region.(*io.Region), ok
}

// OpenRegion opens a region file at the given region X and Z in the given path.
// OpenRegion creates a region file if it did not yet exist.
func (provider *Anvil) OpenRegion(regionX, regionZ int32, path string) {
	var region, _ = io.OpenRegion(path)
	provider.regions.Store(provider.GetChunkIndex(regionX, regionZ), region)
}

// Close closes the provider and saves all chunks.
func (provider *Anvil) Close(async bool) {
	if async {
		go func() {
			provider.regions.Range(func(index, region interface{}) bool {
				region.(*io.Region).Close(true)
				provider.regions.Delete(index)
				return true
			})
		}()
	} else {
		provider.regions.Range(func(index, region interface{}) bool {
			region.(*io.Region).Close(true)
			provider.regions.Delete(index)
			return true
		})
	}
}

// Save saves all regions in the provider.
func (provider *Anvil) Save() {
	go func() {
		provider.regions.Range(func(index, region interface{}) bool {
			region.(*io.Region).Save()
			return true
		})
	}()
}
