package blocks

import "errors"

// Manager manages blocks and has utility functions for registering those.
type Manager map[byte]func(data byte) Block

var UnregisteredBlock = errors.New("block is not registered")

// NewManager returns a new blocks manager.
func NewManager() Manager {
	return Manager{}
}

// Register registers a new block function for the given block ID.
// Register overwrites any blocks that might have been previously registered on the ID.
func (manager Manager) Register(blockId byte, blockFunc func(data byte) Block) {
	manager[blockId] = blockFunc
}

// Deregister deregisters the block function with the given block ID.
func (manager Manager) Deregister(blockId byte) {
	delete(manager, blockId)
}

// IsRegistered checks if a block function with the given block ID is registered.
func (manager Manager) IsRegistered(blockId byte) bool {
	var _, ok = manager[blockId]
	return ok
}

// Get returns a block by its block ID and block data.
// Returns an error if a block with the given block ID was not registered.
func (manager Manager) Get(blockId byte, blockData byte) (Block, error) {
	if !manager.IsRegistered(blockId) {
		return nil, UnregisteredBlock
	}
	return manager[blockId](blockData), nil
}
