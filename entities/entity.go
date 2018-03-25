package entities

import (
	"errors"
	"github.com/golang/geo/r3"
	"github.com/irmine/gomine/net/packets"
	"github.com/irmine/gomine/net/protocol"
	"github.com/irmine/gomine/utils"
	"github.com/irmine/gonbt"
	"github.com/irmine/worlds"
	"github.com/irmine/worlds/chunks"
	"github.com/irmine/worlds/entities/data"
	"math"
	"sync"
)

// Viewer is a viewer of an entity.
// Entities can be sent to the viewer and removed from the viewer.
type Viewer interface {
	chunks.Viewer
	SendAddEntity(protocol.AddEntityEntry)
	SendAddPlayer(utils.UUID, int32, protocol.AddPlayerEntry)
	SendPacket(packet packets.IPacket)
	SendRemoveEntity(int64)
}

// Entity is a movable object in a dimension.
type Entity struct {
	entityType   EntityType
	attributeMap data.AttributeMap

	Position r3.Vector
	Rotation data.Rotation
	Motion   r3.Vector
	OnGround bool

	Dimension *worlds.Dimension

	NameTag string

	runtimeId uint64
	closed    bool

	nbt *gonbt.Compound

	mutex      sync.RWMutex
	EntityData map[uint32][]interface{}
	SpawnedTo  map[utils.UUID]Viewer
}

// UnloadedChunkMove gets returned when the location passed in SetPosition is in an unloaded chunk.
var UnloadedChunkMove = errors.New("tried to move entity in unloaded chunk")

// New returns a new entity by the given entity type.
func New(entityType EntityType) *Entity {
	ent := Entity{
		entityType,
		data.NewAttributeMap(),
		r3.Vector{},
		data.Rotation{},
		r3.Vector{},
		false,
		nil,
		"",
		0,
		true,
		gonbt.NewCompound("", make(map[string]gonbt.INamedTag)),
		sync.RWMutex{},
		make(map[uint32][]interface{}),
		make(map[utils.UUID]Viewer),
	}
	return &ent
}

// GetNameTag returns the name tag of this entity.
func (entity *Entity) GetNameTag() string {
	return entity.NameTag
}

// SetNameTag sets the name tag of this entity.
func (entity *Entity) SetNameTag(nameTag string) {
	entity.NameTag = nameTag
}

// GetAttributeMap returns the attribute map of this entity.
func (entity *Entity) GetAttributeMap() data.AttributeMap {
	return entity.attributeMap
}

// SetAttributeMap sets the attribute map of this entity.
func (entity *Entity) SetAttributeMap(attMap data.AttributeMap) {
	entity.attributeMap = attMap
}

// GetEntityData returns the entity data map.
func (entity *Entity) GetEntityData() map[uint32][]interface{} {
	return entity.EntityData
}

// GetPosition returns the current position of this entity.
func (entity *Entity) GetPosition() r3.Vector {
	return entity.Position
}

// SetPosition sets the position of this entity
func (entity *Entity) SetPosition(v r3.Vector) error {
	var newChunkX = int32(math.Floor(float64(v.X))) >> 4
	var newChunkZ = int32(math.Floor(float64(v.Z))) >> 4

	var oldChunk = entity.GetChunk()
	var newChunk, ok = entity.Dimension.GetChunk(newChunkX, newChunkZ)
	if !ok {
		return UnloadedChunkMove
	}

	entity.Position = v

	if oldChunk != newChunk {
		newChunk.AddEntity(entity)
		entity.SpawnToAll()
		oldChunk.RemoveEntity(entity.runtimeId)
	}
	return nil
}

// IsOnGround checks if the entity is on the ground.
func (entity *Entity) IsOnGround() bool {
	return entity.OnGround
}

// GetChunk returns the chunk this entity is currently in.
func (entity *Entity) GetChunk() *chunks.Chunk {
	var x = int32(math.Floor(float64(entity.Position.X))) >> 4
	var z = int32(math.Floor(float64(entity.Position.Z))) >> 4
	var chunk, _ = entity.Dimension.GetChunk(x, z)
	return chunk
}

// GetViewers returns all players that have the chunk loaded in which this entity is.
func (entity *Entity) GetViewers() map[utils.UUID]Viewer {
	return entity.SpawnedTo
}

// AddViewer adds a viewer to this entity.
func (entity *Entity) AddViewer(viewer Viewer) {
	entity.mutex.Lock()
	entity.SpawnedTo[viewer.GetUUID()] = viewer
	entity.mutex.Unlock()
}

// RemoveViewer removes a viewer from this entity.
func (entity *Entity) RemoveViewer(viewer Viewer) {
	entity.mutex.Lock()
	delete(entity.SpawnedTo, viewer.GetUUID())
	entity.mutex.Unlock()
}

// GetDimension returns the dimension of this entity.
func (entity *Entity) GetDimension() *worlds.Dimension {
	return entity.Dimension
}

// SetDimension sets the dimension of the entity.
func (entity *Entity) SetDimension(v interface {
	GetChunk(int32, int32) (*chunks.Chunk, bool)
}) {
	entity.Dimension = v.(*worlds.Dimension)
}

// GetRotation returns the current rotation of this entity.
func (entity *Entity) GetRotation() data.Rotation {
	return entity.Rotation
}

// SetRotation sets the rotation of this entity.
func (entity *Entity) SetRotation(v data.Rotation) {
	entity.Rotation = v
}

// GetMotion returns the motion of this entity.
func (entity *Entity) GetMotion() r3.Vector {
	return entity.Motion
}

// SetMotion sets the motion of this entity.
func (entity *Entity) SetMotion(v r3.Vector) {
	entity.Motion = v
}

// GetRuntimeId returns the runtime ID of the entity.
func (entity *Entity) GetRuntimeId() uint64 {
	return entity.runtimeId
}

// SetRuntimeId sets the runtime ID of the entity.
// SetRuntimeId should not be used by plugins.
func (entity *Entity) SetRuntimeId(id uint64) {
	entity.runtimeId = id
}

// GetUniqueId returns the unique ID of this entity.
// NOTE: This is currently unimplemented, and returns the runtime ID.
func (entity *Entity) GetUniqueId() int64 {
	return int64(entity.runtimeId)
}

// GetEntityType returns the entity type of this entity.
func (entity *Entity) GetEntityType() uint32 {
	return uint32(entity.entityType)
}

// IsClosed checks if the entity is closed and not to be used anymore.
func (entity *Entity) IsClosed() bool {
	return entity.closed
}

// Close closes the entity making it unable to be used.
func (entity *Entity) Close() {
	entity.closed = true
	entity.DespawnFromAll()

	entity.Dimension = nil
	entity.SpawnedTo = nil
}

// GetHealth returns the health points of this entity.
func (entity *Entity) GetHealth() float32 {
	return entity.attributeMap.GetAttribute(data.AttributeHealth).GetValue()
}

// SetHealth sets the health points of this entity.
func (entity *Entity) SetHealth(health float32) {
	entity.attributeMap.GetAttribute(data.AttributeHealth).SetValue(health)
}

// Kill kills the entity.
func (entity *Entity) Kill() {
	entity.SetHealth(0)
}

// SpawnTo spawns this entity to the given player.
func (entity *Entity) SpawnTo(viewer Viewer) {
	if entity.IsClosed() {
		return
	}
	entity.AddViewer(viewer)
	viewer.SendAddEntity(entity)
}

// DespawnFrom despawns this entity from the given player.
func (entity *Entity) DespawnFrom(viewer Viewer) {
	entity.RemoveViewer(viewer)
	viewer.SendRemoveEntity(entity.GetUniqueId())
}

// DespawnFromAll despawns this entity from all viewers.
func (entity *Entity) DespawnFromAll() {
	for _, viewer := range entity.SpawnedTo {
		entity.DespawnFrom(viewer)
	}
}

// SpawnToAll spawns this entity to all players.
func (entity *Entity) SpawnToAll() {
	for _, v := range entity.GetChunk().GetViewers() {
		var (
			viewer Viewer
			ok     bool
		)
		if viewer, ok = v.(Viewer); !ok {
			continue
		}
		if _, ok := entity.SpawnedTo[viewer.GetUUID()]; !ok {
			entity.SpawnTo(viewer)
		}
	}
}

// GetNBT returns the NBT data of the entity.
func (entity *Entity) GetNBT() *gonbt.Compound {
	return entity.nbt
}

// SetNBT sets the NBT data of the entity.
func (entity *Entity) SetNBT(nbt *gonbt.Compound) {
	entity.nbt = nbt
}

// Tick ticks the entity.
func (entity *Entity) Tick() {

}
