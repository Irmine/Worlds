package worlds

import (
	"os"
	"sync"
)

// Level is a struct that manages an unlimited set of dimensions.
// Every level has its own set of game rules.
type Level struct {
	name             string
	serverPath       string
	defaultDimension *Dimension

	currentTick int64

	mutex      sync.RWMutex
	dimensions map[string]*Dimension
	gameRules  map[GameRuleName]*GameRule
}

// NewLevel returns a new level with the given level name and server path.
// World data will be generated in: `serverPath/worlds/`
func NewLevel(levelName string, serverPath string) *Level {
	var level = &Level{levelName, serverPath, nil, 0, sync.RWMutex{}, make(map[string]*Dimension), make(map[GameRuleName]*GameRule)}
	os.MkdirAll(serverPath+"worlds/"+levelName, 0700)

	level.initializeGameRules()
	return level
}

// GetGameRule returns a game rule with the given name.
func (level *Level) GetGameRule(gameRule GameRuleName) *GameRule {
	level.mutex.RLock()
	defer level.mutex.RUnlock()
	return level.gameRules[gameRule]
}

// GetGameRules returns all game rules of the level in a name => game rule map.
func (level *Level) GetGameRules() map[GameRuleName]*GameRule {
	level.mutex.RLock()
	defer level.mutex.RUnlock()
	return level.gameRules
}

// AddGameRule adds the given game rule to the level.
func (level *Level) AddGameRule(rule *GameRule) {
	level.mutex.Lock()
	level.gameRules[rule.GetName()] = rule
	level.mutex.Unlock()
}

// GetName returns the name of the level.
func (level *Level) GetName() string {
	return level.name
}

// GetDimensions returns all dimensions of the level in a name => dimension map.
func (level *Level) GetDimensions() map[string]*Dimension {
	level.mutex.RLock()
	defer level.mutex.RUnlock()
	return level.dimensions
}

// DimensionExists checks if a dimension with the given name exists in the level.
func (level *Level) DimensionExists(name string) bool {
	level.mutex.RLock()
	var _, exists = level.dimensions[name]
	level.mutex.RUnlock()
	return exists
}

// AddDimension adds a dimension to the level, overwriting it if needed.
// Returns a bool indicating if a dimension got overwritten.
func (level *Level) AddDimension(dimension *Dimension) bool {
	var exists = level.DimensionExists(dimension.GetName())
	level.mutex.Lock()
	level.dimensions[dimension.GetName()] = dimension
	level.mutex.Unlock()
	return exists
}

// SetDefaultDimension sets the default dimension of the level.
// Also adds the level if had not yet been added.
func (level *Level) SetDefaultDimension(dimension *Dimension) {
	level.AddDimension(dimension)

	level.defaultDimension = dimension
}

// GetDefaultDimension returns the default dimension of the level.
func (level *Level) GetDefaultDimension() *Dimension {
	return level.defaultDimension
}

// RemoveDimension removes a dimension in the level with the given name.
// It returns a bool indicating success.
func (level *Level) RemoveDimension(name string) bool {
	if !level.DimensionExists(name) {
		return false
	}
	level.mutex.Lock()
	delete(level.dimensions, name)
	level.mutex.Unlock()
	return true
}

// GetDimension returns a dimension by its name and a bool indicating success.
func (level *Level) GetDimension(name string) (*Dimension, bool) {
	if !level.DimensionExists(name) {
		return nil, false
	}
	return level.dimensions[name], true
}

// Tick ticks the level, ticking all dimensions and their contents.
func (level *Level) Tick() {
	level.currentTick++
	for _, dimension := range level.dimensions {
		dimension.Tick()
	}
}

// GetCurrentTick returns the amount of ticks this level has had.
func (level *Level) GetCurrentTick() int64 {
	return level.currentTick
}

// initializeGameRules initializes all game rules of the level, setting them to their default values.
func (level *Level) initializeGameRules() {
	level.AddGameRule(NewGameRule(GameRuleCommandBlockOutput, true))
	level.AddGameRule(NewGameRule(GameRuleDoDaylightCycle, true))
	level.AddGameRule(NewGameRule(GameRuleDoEntityDrops, true))
	level.AddGameRule(NewGameRule(GameRuleDoFireTick, true))
	level.AddGameRule(NewGameRule(GameRuleDoMobLoot, true))
	level.AddGameRule(NewGameRule(GameRuleDoMobSpawning, true))
	level.AddGameRule(NewGameRule(GameRuleDoTileDrops, true))
	level.AddGameRule(NewGameRule(GameRuleDoWeatherCycle, true))
	level.AddGameRule(NewGameRule(GameRuleDrowningDamage, true))
	level.AddGameRule(NewGameRule(GameRuleFallDamage, true))
	level.AddGameRule(NewGameRule(GameRuleFireDamage, true))
	level.AddGameRule(NewGameRule(GameRuleKeepInventory, false))
	level.AddGameRule(NewGameRule(GameRuleMobGriefing, true))
	level.AddGameRule(NewGameRule(GameRuleNaturalRegeneration, true))
	level.AddGameRule(NewGameRule(GameRulePvp, true))
	level.AddGameRule(NewGameRule(GameRuleSendCommandFeedback, true))
	level.AddGameRule(NewGameRule(GameRuleShowCoordinates, true))
	level.AddGameRule(NewGameRule(GameRuleRandomTickSpeed, uint32(3)))
	level.AddGameRule(NewGameRule(GameRuleTntExplodes, true))
}
