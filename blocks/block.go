package blocks

import (
	"github.com/irmine/gonbt"
)

type Block interface {
	GetId() byte
	GetData() byte
	SetData(byte)
	GetNBT() *gonbt.Compound
	SetNBT(*gonbt.Compound)
}

type BlockInstance struct {
	*BlockState
	nbt *gonbt.Compound
}

func New(state *BlockState) *BlockInstance {
	return &BlockInstance{state, nil}
}

func (base *BlockInstance) GetNBT() *gonbt.Compound {
	return base.nbt
}

func (base *BlockInstance) SetNBT(nbt *gonbt.Compound) {
	base.nbt = nbt
}
