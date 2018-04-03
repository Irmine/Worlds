package blocks

// Block positions are rounded block positions.
// Block positions do not require floating point
// numbers, and therefore have a different position.
type Position struct {
	// X and Z of the block position are int32.
	// The X, Y and Z values are in the minimal corner of a block.
	X, Z int32
	// The Y of the block position is uint32.
	// Block positions can never be below zero.
	Y uint32
}

// NewPosition returns a new position with X, Y and Z.
func NewPosition(x int32, y uint32, z int32) Position {
	return Position{X: x, Y: y, Z: z}
}
