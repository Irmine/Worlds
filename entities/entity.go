package entities

import (
	"errors"
	"github.com/golang/geo/r3"
	"github.com/irmine/gonbt"
	"github.com/irmine/worlds"
	"github.com/irmine/worlds/chunks"
	"math"
	"sync"
)

type EntityViewer interface {
	GetRuntimeId() uint64
	IsClosed() bool
	SendAddEntity(*Entity)
	SendRemoveEntity(*Entity)
}

type Entity struct {
	entityType   EntityType
	attributeMap AttributeMap

	Position r3.Vector
	Rotation Rotation
	Motion   r3.Vector

	Level     *worlds.Level
	Dimension *worlds.Dimension

	NameTag   string
	SpawnedTo map[uint64]EntityViewer

	mutex sync.Mutex

	EntityData map[uint32][]interface{}

	runtimeId uint64
	closed    bool
}

type Rotation struct {
	Yaw, Pitch float64
}

func New(entityType EntityType) *Entity {
	ent := Entity{
		entityType,
		NewAttributeMap(),
		r3.Vector{},
		Rotation{},
		r3.Vector{},
		nil,
		nil,
		"",
		make(map[uint64]EntityViewer),
		sync.Mutex{},
		make(map[uint32][]interface{}),
		0,
		true,
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
func (entity *Entity) GetAttributeMap() AttributeMap {
	return entity.attributeMap
}

// SetAttributeMap sets the attribute map of this entity.
func (entity *Entity) SetAttributeMap(attMap AttributeMap) {
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
		return errors.New("entity tried moving in unloaded chunk")
	}

	entity.Position = v

	if oldChunk != newChunk {
		newChunk.AddEntity(entity)
		entity.SpawnToAll()
		oldChunk.RemoveEntity(entity.runtimeId)
	}
	return nil
}

// GetChunk returns the chunk this entity is currently in.
func (entity *Entity) GetChunk() *chunks.Chunk {
	var x = int32(math.Floor(float64(entity.Position.X))) >> 4
	var z = int32(math.Floor(float64(entity.Position.Z))) >> 4
	var chunk, _ = entity.Dimension.GetChunk(x, z)
	return chunk
}

// GetViewers returns all players that have the chunk loaded in which this entity is.
func (entity *Entity) GetViewers() map[uint64]EntityViewer {
	return entity.SpawnedTo
}

// AddViewer adds a viewer to this entity.
func (entity *Entity) AddViewer(viewer EntityViewer) {
	entity.mutex.Lock()
	entity.SpawnedTo[viewer.GetRuntimeId()] = viewer
	entity.mutex.Unlock()
}

// RemoveViewer removes a viewer from this entity.
func (entity *Entity) RemoveViewer(viewer EntityViewer) {
	entity.mutex.Lock()
	delete(entity.SpawnedTo, viewer.GetRuntimeId())
	entity.mutex.Unlock()
}

// GetLevel returns the level of this entity.
func (entity *Entity) GetLevel() *worlds.Level {
	return entity.Level
}

// SetLevel sets the level of this entity.
func (entity *Entity) SetLevel(v *worlds.Level) {
	entity.Level = v
}

// GetDimension returns the dimension of this entity.
func (entity *Entity) GetDimension() *worlds.Dimension {
	return entity.Dimension
}

// SetDimension sets the dimension of the entity.
func (entity *Entity) SetDimension(v *worlds.Dimension) {
	entity.Dimension = v
}

// GetRotation returns the current rotation of this entity.
func (entity *Entity) GetRotation() Rotation {
	return entity.Rotation
}

// SetRotation sets the rotation of this entity.
func (entity *Entity) SetRotation(v Rotation) {
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

// GetRuntimeId returns the runtime ID of this entity.
func (entity *Entity) GetRuntimeId() uint64 {
	return entity.runtimeId
}

// GetUniqueId returns the unique ID of this entity.
// NOTE: This is currently unimplemented, and returns the runtime ID.
func (entity *Entity) GetUniqueId() int64 {
	return int64(entity.runtimeId)
}

// GetEntityId returns the entity ID of this entity.
func (entity *Entity) GetEntityId() uint32 {
	return 0
}

// IsClosed checks if the entity is closed and not to be used anymore.
func (entity *Entity) IsClosed() bool {
	return entity.closed
}

// Close closes the entity making it unable to be used.
func (entity *Entity) Close() {
	entity.closed = true
	entity.Level = nil
	entity.Dimension = nil
	entity.SpawnedTo = nil
}

// GetHealth returns the health points of this entity.
func (entity *Entity) GetHealth() float32 {
	return entity.attributeMap.GetAttribute(AttributeHealth).GetValue()
}

// SetHealth sets the health points of this entity.
func (entity *Entity) SetHealth(health float32) {
	entity.attributeMap.GetAttribute(AttributeHealth).SetValue(health)
}

// Kill kills the entity.
func (entity *Entity) Kill() {
	entity.SetHealth(0)
}

// SpawnTo spawns this entity to the given player.
func (entity *Entity) SpawnTo(viewer EntityViewer) {
	if viewer.IsClosed() {
		return
	}
	if entity.GetRuntimeId() == viewer.GetRuntimeId() {
		return
	}
	entity.AddViewer(viewer)
	viewer.SendAddEntity(entity)
}

// DespawnFrom despawns this entity from the given player.
func (entity *Entity) DespawnFrom(viewer EntityViewer) {
	if viewer.IsClosed() {
		return
	}
	entity.RemoveViewer(viewer)
	viewer.SendRemoveEntity(entity)
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
		if v.GetRuntimeId() != entity.GetRuntimeId() {
			var (
				viewer EntityViewer
				ok     bool
			)
			if viewer, ok = v.(EntityViewer); !ok {
				continue
			}
			if _, ok := entity.SpawnedTo[viewer.GetRuntimeId()]; !ok {
				entity.SpawnTo(viewer)
			}
		}
	}
}

// GetSaveData returns the NBT save data of the entity.
func (entity *Entity) GetSaveData() *gonbt.Compound {
	return nil
}

// Tick ticks the entity.
func (entity *Entity) Tick() {
	for runtimeId, player := range entity.GetViewers() {
		if player.IsClosed() {
			delete(entity.SpawnedTo, runtimeId)
		}
	}
}
