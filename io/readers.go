package io

import (
	"gomine/interfaces"
	"libraries/nbt"
)

/**
 * Returns a new Anvil chunk from the given NBT compound.
 */
func GetAnvilChunkFromNBT(compound *GoNBT.Compound) interfaces.IChunk {
	var level = compound.GetCompound("Level")
	var chunk = NewChunk(level.GetInt("xPos", 0), level.GetInt("zPos", 0))
	chunk.LightPopulated = getBool(level.GetByte("LightPopulated", 0))
	chunk.TerrainPopulated = getBool(level.GetByte("TerrainPopulated", 0))
	chunk.biomes = level.GetByteArray("Biomes", make([]byte, 256))
	chunk.InhabitedTime = level.GetLong("InhabitedTime", 0)
	chunk.LastUpdate = level.GetLong("LastUpdate", 0)
	var heightMap = [256]int16{}
	for i, b := range level.GetByteArray("HeightMap", make([]byte, 256)) {
		heightMap[i] = int16(b)
	}
	chunk.heightMap = heightMap

	var sections = level.GetList("Sections", GoNBT.TAG_Compound)
	if sections == nil {
		return chunk
	}
	for _, comp := range sections.GetTags() {
		section := comp.(*GoNBT.Compound)
		subChunk := NewSubChunk()
		subChunk.BlockLight = reorderNibbleArray(section.GetByteArray("BlockLight", make([]byte, 2048)))
		//subChunk.BlockData = (section.GetByteArray("Data", make([]byte, 2048)))
		subChunk.SkyLight = reorderNibbleArray(section.GetByteArray("SkyLight", make([]byte, 2048)))
		subChunk.BlockIds = reorderBlocks(section.GetByteArray("Blocks", make([]byte, 4096)))

		chunk.subChunks[section.GetByte("Y", 0)] = subChunk
	}

	return chunk
}

func reorderBlocks(blocks []byte) []byte {
	var data = make([]byte, 4096)
	var i = 0
	for x := 0; x < 16; x++ {
		var zM = x + 256
		for z := x; z < zM; z += 16 {
			var yM = z + 4096
			for y := z; y < yM; y += 256 {
				data[i] = blocks[y]
				i++
			}
		}
	}
	return data
}

func reorderNibbleArray(arr []byte) []byte {
	var data = make([]byte, 2048)
	var i = 0
	for x := 0; x < 8; x++ {
		for z := 0; z < 16; z++ {
			var zx = (z << 3) | x
			for y := 0; y < 8; y++ {
				var j = (y << 8) | zx
				var j80 = j | 0x80
				if arr[j] != 0 || arr[j80] != 0 {
					data[i] = (arr[j80] << 4) | (arr[j] & 0x0f)
					data[i|80] = (arr[j] >> 4) | (arr[j80] & 0xf0)
				}
				i++
			}
		}
		i += 128
	}
	return data
}

func getBool(value byte) bool {
	if value > 0 {
		return true
	}
	return false
}
