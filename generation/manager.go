package generation

import "errors"

// Manager is a map used to manage generators.
type Manager map[string]Generator

// UnregisteredGenerator gets returned when trying to get a generator that is not registered.
var UnregisteredGenerator = errors.New("generator is not registered")

// NewManager returns a new manager.
func NewManager() Manager {
	return Manager{}
}

// Register registers the given generator.
func (manager Manager) Register(generator Generator) {
	manager[generator.GetName()] = generator
}

// Deregister deregisters a generator with the given name.
func (manager Manager) Deregister(name string) {
	delete(manager, name)
}

// Get attempts to return a registered generator with the given name.
// Returns UnregisteredGenerator if none could be found.
func (manager Manager) Get(name string) (Generator, error) {
	if !manager.IsRegistered(name) {
		return nil, UnregisteredGenerator
	}
	return manager[name], nil
}

// Exists checks if a generator with the given name exists.
func (manager Manager) IsRegistered(name string) bool {
	var _, ok = manager[name]
	return ok
}
