package blocks

import "github.com/irmine/gonbt"

type BlockState struct {
	name      string
	runtimeId int32
	LegacyId  byte
	data      byte
}

// NewBlockState returns a new block state with the given name, ID and data.
func NewBlockState(name string, runtimeId int32, LegacyId, data byte) *BlockState {
	return &BlockState{name, runtimeId,  LegacyId, data}
}

// GetName returns the Minecraft name of the block state.
func (state *BlockState) GetName() string {
	return state.name
}

// GetRuntimeId returns the runtime ID of the block state.
func (state *BlockState) GetRuntimeId() int32 {
	return state.runtimeId
}

// GetData returns the legacy id of the block state.
func (state *BlockState) GetId() byte {
	return state.LegacyId
}

// SetData sets the block state's legacy id.
func (state *BlockState) SetId(legacyId byte) {
	state.LegacyId = legacyId
}

// GetData returns the data of the block state.
func (state *BlockState) GetData() byte {
	return state.data
}

// SetData sets the block state's data.
func (state *BlockState) SetData(data byte) {
	state.data = data
}

// GetPersistentId returns the persistent ID NBT of the block state.
// This persistent ID remains the same through every session.
func (state *BlockState) GetPersistentId() *gonbt.Compound {
	return gonbt.NewCompound("", map[string]gonbt.INamedTag{
		"name": gonbt.NewString("name", state.GetName()),
		"val":  gonbt.NewInt("val", int32(state.data)),
	})
}
