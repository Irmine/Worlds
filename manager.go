package worlds

import (
	"errors"
	"github.com/irmine/worlds/generation"
	"os"
	"sync"
)

// Manager is a struct managing all levels and provides helper functions.
type Manager struct {
	serverPath       string
	generatorManager generation.Manager

	defaultLevel *Level
	mutex        sync.RWMutex
	levels       map[string]*Level
}

// NewManager returns a new worlds manager.
// The manager will create its content inside of the `serverPath/worlds/` folder.
func NewManager(serverPath string) *Manager {
	os.MkdirAll(serverPath+"/worlds", 0700)
	return &Manager{serverPath, generation.NewManager(), nil, sync.RWMutex{}, make(map[string]*Level)}
}

// GetLoadedLevels returns all loaded levels of the manager in a name => level map.
func (manager *Manager) GetLevels() map[string]*Level {
	return manager.levels
}

// GetGenerationManager returns the generator manager of the level manager.
func (manager *Manager) GetGenerationManager() generation.Manager {
	return manager.generatorManager
}

// IsLevelLoaded checks if a level is loaded with the given name.
func (manager *Manager) IsLevelLoaded(levelName string) bool {
	manager.mutex.RLock()
	var _, ok = manager.levels[levelName]
	manager.mutex.RUnlock()
	return ok
}

// IsLevelGenerated checks if a level with the given name has been generated.
func (manager *Manager) IsLevelGenerated(levelName string) bool {
	if manager.IsLevelLoaded(levelName) {
		return true
	}
	var path = manager.serverPath + "worlds/" + levelName
	var _, err = os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

// LoadLevel loads a level with the given name, and returns a bool indicating success.
func (manager *Manager) LoadLevel(levelName string) bool {
	if !manager.IsLevelGenerated(levelName) {
		// manager.GenerateLevel(level) We need file writing for this. TODO.
	}
	if manager.IsLevelLoaded(levelName) {
		return false
	}
	manager.mutex.Lock()
	manager.levels[levelName] = NewLevel(levelName, manager.serverPath)
	manager.mutex.Unlock()
	return true
}

// GetDefaultLevel returns the default level of the manager.
func (manager *Manager) GetDefaultLevel() *Level {
	return manager.defaultLevel
}

// SetDefaultLevel sets the given level as default, and adds it if needed.
func (manager *Manager) SetDefaultLevel(level *Level) {
	manager.mutex.Lock()
	manager.levels[level.GetName()] = level
	manager.mutex.Unlock()
	manager.defaultLevel = level
}

// GetLevel returns a level by its name, or an error if something went wrong.
func (manager *Manager) GetLevel(name string) (*Level, error) {
	if !manager.IsLevelGenerated(name) {
		return nil, errors.New("level with given name is not generated")
	}
	if !manager.IsLevelLoaded(name) {
		return nil, errors.New("level with given name is not loaded")
	}

	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return manager.levels[name], nil
}

// Tick ticks all levels managed by the Manager.
func (manager *Manager) Tick() {
	for _, level := range manager.levels {
		level.Tick()
	}
}

// Close closes all levels and their dimensions.
func (manager *Manager) Close() {
	for _, level := range manager.levels {
		for _, dimension := range level.GetDimensions() {
			dimension.Close(false)
		}
	}
}

// Save saves all levels and their dimensions.
func (manager *Manager) Save() {
	for _, level := range manager.levels {
		for _, dimension := range level.GetDimensions() {
			dimension.Save()
		}
	}
}
