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

type Base struct {
	id   byte
	data byte
	nbt  *gonbt.Compound
}

func New(id byte, data byte) *Base {
	return &Base{id, data, nil}
}

func (base *Base) GetId() byte {
	return base.id
}

func (base *Base) GetData() byte {
	return base.data
}

func (base *Base) SetData(data byte) {
	base.data = data
}

func (base *Base) GetNBT() *gonbt.Compound {
	return base.nbt
}

func (base *Base) SetNBT(nbt *gonbt.Compound) {
	base.nbt = nbt
}
