package blocks

import "github.com/golang/geo/r3"

// Batch is a batch containing multiple blocks.
// They can be used to set a batch of blocks to a dimension.
type Batch map[r3.Vector]Block

// NewBatch returns a new block batch with the given blocks.
func NewBatch(blocks map[r3.Vector]Block) Batch {
	return Batch(blocks)
}

// AddBlock adds the given block with position to the batch.
func (batch Batch) AddBlock(vector r3.Vector, block Block) {
	batch[vector] = block
}

// AddBlocks adds the given block map to the batch.
func (batch Batch) AddBlocks(blocks map[r3.Vector]Block) {
	for v, b := range blocks {
		batch[v] = b
	}
}

// Merge merges the batch with another batch.
func (batch Batch) Merge(batch2 Batch) {
	for v, b := range batch2 {
		batch[v] = b
	}
}
