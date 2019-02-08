package entities

import (
	"errors"
	"github.com/golang/geo/r3"
	"github.com/google/uuid"
	"github.com/irmine/gomine/net/packets"
	"github.com/irmine/gomine/net/protocol"
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
	SendAddPlayer(uuid.UUID, protocol.AddPlayerEntry)
	SendPacket(packet packets.IPacket)
	SendRemoveEntity(int64)
	SendMoveEntity(uint64, r3.Vector, data.Rotation, byte, bool)
	SendMovePlayer(uint64, r3.Vector, data.Rotation, byte, bool, uint64)
	SendSetEntityData(uint64, map[uint32][]interface{})
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

	ridingId  uint64
	runtimeId uint64
	closed    bool

	nbt *gonbt.Compound

	mutex      sync.RWMutex
	entityData map[uint32][]interface{}
	updatedEntityData map[uint32][]interface{}
	SpawnedTo  map[uuid.UUID]Viewer

	HasEntityDataUpdate bool
	HasMovementUpdate bool
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
		0,
		false,
		gonbt.NewCompound("", make(map[string]gonbt.INamedTag)),
		sync.RWMutex{},
		make(map[uint32][]interface{}),
		make(map[uint32][]interface{}),
		make(map[uuid.UUID]Viewer),
		true,
		false,
	}

	//ent.SetEntityDataFlag(data.EntityDataIdFlags, data.EntityDataLong, 0)

	ent.SetEntityProperty(data.EntityDataAffectedByGravity, true)
	ent.SetEntityDataFlag(data.EntityDataMaxAir, data.EntityDataShort, 400)

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

// UpdateEntityData tells the entity that there is a data updated that needs to be send
func (entity *Entity) UpdateEntityData() {
	entity.HasEntityDataUpdate = true
}

// EntityDataFlagExists returns whether a flag exists or not
func (entity *Entity) EntityDataFlagExists(flagId uint32) bool {
	_, ok := entity.entityData[flagId]
	return ok
}

// SetEntityDataFlag sets a property id, flag id, and a value to the entity's data
func (entity *Entity) SetEntityDataFlag(propId, flagId uint32, value interface{}) {
	entity.entityData[propId], entity.updatedEntityData[propId] = []interface{}{flagId, value}, []interface{}{flagId, value}
	entity.UpdateEntityData()
}

// RemoveEntityDataFlag removes the entity data from a given property id
func (entity *Entity) RemoveEntityDataFlag(propId uint32) {
	delete(entity.entityData, propId)
	delete(entity.updatedEntityData, propId)
	entity.UpdateEntityData()
}

// GetDataFlag returns the values of the entity's data from a given property id
// if there is no data found it will return negative integers with a -1 value
func (entity *Entity) GetDataFlag(propId uint32) (v []interface{}) {
	if v, ok := entity.entityData[propId]; ok {
		return v
	}
	return []interface{}{-1, -1}
}

// GetEntityData returns the entity data map.
func (entity *Entity) GetEntityData() map[uint32][]interface{} {
	return entity.entityData
}

// GetEntityData shifts and returns updated entity data for sending
func (entity *Entity) GetUpdatedEntityData() map[uint32][]interface{} {
	var entityData = entity.updatedEntityData
	entity.updatedEntityData = make(map[uint32][]interface{})
	return entityData
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
func (entity *Entity) GetViewers() map[uuid.UUID]Viewer {
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

// GetRidingId returns the runtime ID of the entity riding.
func (entity *Entity) GetRidingId() uint64 {
	return entity.ridingId
}

// SetRidingId sets the runtime ID of the entity riding.
// SetRidingId should not be used by plugins.
func (entity *Entity) SetRidingId(id uint64) {
	entity.ridingId = id
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
	return entity.attributeMap.GetAttribute(data.AttributeHealth).Value
}

// SetHealth sets the health points of this entity.
func (entity *Entity) SetHealth(health float32) {
	entity.attributeMap.GetAttribute(data.AttributeHealth).Value = health
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

// Sets a generic data flag by it's flag id, if value is true
// it will set the flag, otherwise it will remove the flag
func (entity *Entity) SetEntityProperty(flagId uint32, value bool) {
	if entity.EntityDataFlagExists(data.EntityDataIdFlags) && !value {
		entity.SetEntityDataFlag(data.EntityDataIdFlags, data.EntityDataLong, int64(1 << flagId) ^ int64(1 << flagId))
	}else{
		entity.SetEntityDataFlag(data.EntityDataIdFlags, data.EntityDataLong, int64(1 << flagId))
	}
}

// Sends base entity data to a certain viewer
func (entity *Entity) SendEntityData(viewer Viewer) {
	viewer.SendSetEntityData(entity.GetRuntimeId(), entity.GetEntityData())
}

// Sends base entity data to all viewers
func (entity *Entity) BroadcastEntityData() {
	for _, viewer := range entity.GetViewers() {
		viewer.SendSetEntityData(entity.GetRuntimeId(), entity.GetEntityData())
	}
}

// Sends updated entity data to a certain viewer
func (entity *Entity) SendUpdatedEntityData(viewer Viewer) {
	viewer.SendSetEntityData(entity.GetRuntimeId(), entity.GetUpdatedEntityData())
}

// Sends updated entity data to all viewers
func (entity *Entity) BroadcastUpdatedEntityData() {
	for _, viewer := range entity.GetViewers() {
		viewer.SendSetEntityData(entity.GetRuntimeId(), entity.GetUpdatedEntityData())
	}
}

// Sends updated entity position and rotation to a certain viewer
func (entity *Entity) SendMovement(viewer Viewer) {
	viewer.SendMoveEntity(entity.runtimeId, entity.Position, entity.Rotation, 0, entity.OnGround)
}

// Sends updated entity position and rotation to all viewers
func (entity *Entity) BroadcastMovement() {
	for _, viewer := range entity.GetViewers() {
		viewer.SendMoveEntity(entity.runtimeId, entity.Position, entity.Rotation, 0, entity.OnGround)
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
	if entity.HasEntityDataUpdate {
		entity.BroadcastUpdatedEntityData()
		entity.HasEntityDataUpdate = false
	}
	if entity.HasMovementUpdate {
		entity.HasMovementUpdate = false
	}
	entity.BroadcastMovement()
}
