package blocks

import "github.com/irmine/gonbt"

type BlockState struct {
	name      string
	runtimeId int32
	data      int32
}

// NewBlockState returns a new block state with the given name, ID and data.
func NewBlockState(name string, runtimeId int32, data int32) *BlockState {
	return &BlockState{name, runtimeId, data}
}

// GetName returns the Minecraft name of the block state.
func (state *BlockState) GetName() string {
	return state.name
}

// GetRuntimeId returns the runtime ID of the block state.
func (state *BlockState) GetRuntimeId() int32 {
	return state.runtimeId
}

// GetData returns the data of the block state.
func (state *BlockState) GetData() int32 {
	return state.data
}

// SetData sets the block state's data.
func (state *BlockState) SetData(data int32) {
	state.data = data
}

// GetPersistentId returns the persistent ID NBT of the block state.
// This persistent ID remains the same through every session.
func (state *BlockState) GetPersistentId() *gonbt.Compound {
	return gonbt.NewCompound("", map[string]gonbt.INamedTag{
		"name": gonbt.NewString("name", state.GetName()),
		"val":  gonbt.NewInt("val", state.data),
	})
}
