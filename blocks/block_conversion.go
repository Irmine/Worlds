package blocks

import (
	"github.com/irmine/binutils"
	"github.com/irmine/gomine/text"
	"gopkg.in/yaml.v2"
)

var runtimeIdsTable []byte
var legacyToRuntimeId = make(map[int]uint32)
var runtimeToLegacyId = make(map[uint32]int)

func registerRuntimeIds() []byte {
	if len(runtimeIdsTable) == 0 {
		var data interface{}
		err := yaml.Unmarshal([]byte(RuntimeIdsTable__), &data)
		if err != nil {
			text.DefaultLogger.Error(err)
			return nil
		}
		stream := binutils.NewStream()
		stream.ResetStream()
		if data2, ok := data.([]interface{}); ok {
			stream.PutUnsignedVarInt(uint32(len(data2)))
			for k, v := range data2 {
				if v2, ok := v.(map[interface{}]interface{}); ok {
					blockId := v2["id"].(int)
					blockData := v2["data"].(int)

					stream.PutString(v2["name"].(string))
					stream.PutLittleShort(int16(blockData))

					legacyToRuntimeId[(blockId << 4) | blockData] = uint32(k)
					runtimeToLegacyId[uint32(k)] = (blockId << 4) | blockData
				}
			}
		}
		runtimeIdsTable = stream.GetBuffer()
	}
	return runtimeIdsTable
}

func GetRuntimeIdsTable() []byte {
	if len(runtimeIdsTable) == 0 {
		registerRuntimeIds()
	}
	return runtimeIdsTable
}

func GetRuntimeId(blockId, blockData int) (uint32, bool) {
	v, ok := legacyToRuntimeId[(blockId << 4) | blockData]
	return v, ok
}

func GetLegacyId(runtimeId uint32) (int, bool) {
	v, ok := runtimeToLegacyId[runtimeId]
	return v, ok
}