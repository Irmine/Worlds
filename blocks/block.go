package blocks

import "github.com/irmine/nbt"

type Block interface {
	GetId() byte
	GetData() byte
	SetData(byte)
	GetNBT() *nbt.Compound
	SetNBT(*nbt.Compound)
}

type Base struct {
	id   byte
	data byte
	nbt  *nbt.Compound
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

func (base *Base) GetNBT() *nbt.Compound {
	return base.nbt
}

func (base *Base) SetNBT(nbt *nbt.Compound) {
	base.nbt = nbt
}
