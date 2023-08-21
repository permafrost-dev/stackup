package scripting

import (
	"sync"

	"github.com/stackup-app/stackup/lib/types"
)

var Registry *ScriptRegistry

func CreateNewScriptzRegistry() *ScriptRegistry {
	if Registry == nil {
		Registry = &ScriptRegistry{
			Items:     sync.Map{},
			Installed: sync.Map{},
		}
	}

	return Registry
}

type ScriptRegistry struct {
	Items     sync.Map
	Installed sync.Map
}

func (sr *ScriptRegistry) Install(ex *types.ScriptExtensionContract) {
	_, installed := sr.Installed.Load((*ex).GetName())
	if !installed {
		sr.Installed.Store((*ex).GetName(), ex)
		(*ex).Install()
	}
}

func (sr *ScriptRegistry) Each(fn func(key string, value *types.ScriptExtensionContract)) {
	sr.Items.Range(func(key, value any) bool {
		k := key.(string)
		v := value.(*types.ScriptExtensionContract)
		fn(k, v)
		return true
	})
}

func (sr *ScriptRegistry) IsInstalled(name string) bool {
	_, found := sr.Installed.Load(name)
	return found
}

func (sr *ScriptRegistry) Add(name string, value *types.ScriptExtension) {
	if !sr.IsInstalled(name) {
		sr.Items.Store(name, value)
	}
}

func (sr *ScriptRegistry) Has(name string) bool {
	_, ok := sr.Items.Load(name)
	return ok
}
