package generation

import "errors"

// Manager is a struct used to manage generators.
type Manager struct {
	generators map[string]Generator
}

var UnregisteredGenerator = errors.New("generator with given name is not registered")

// NewManager returns a new manager.
func NewManager() *Manager {
	return &Manager{make(map[string]Generator)}
}

// Register registers the given generator.
func (manager *Manager) Register(generator Generator) {
	manager.generators[generator.GetName()] = generator
}

// Deregister deregisters a generator with the given name.
func (manager *Manager) Deregister(name string) {
	delete(manager.generators, name)
}

// Get attempts to return a registered generator with the given name.
// Returns UnregisteredGenerator if none could be found.
func (manager *Manager) Get(name string) (Generator, error) {
	if !manager.Exists(name) {
		return nil, UnregisteredGenerator
	}
	return manager.generators[name], nil
}

// Exists checks if a generator with the given name exists.
func (manager *Manager) Exists(name string) bool {
	var _, ok = manager.generators[name]
	return ok
}
