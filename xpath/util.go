package xpath

import (
	"sync"
)

var (
	contextMap   = make(map[interface{}]VariableScope)
	contextMutex sync.Mutex
)

// GetScope retrieves a VariableScope for a context.
func GetScope(ctxt interface{}) VariableScope {
	contextMutex.Lock()
	context := contextMap[ctxt]
	contextMutex.Unlock()
	return context
}

// SetScope associates a VariableScope with a context.
func SetScope(ctxt interface{}, v VariableScope) {
	contextMutex.Lock()
	contextMap[ctxt] = v
	contextMutex.Unlock()
}

// ClearScope removes a VariableScope association.
func ClearScope(ctxt interface{}) {
	contextMutex.Lock()
	delete(contextMap, ctxt)
	contextMutex.Unlock()
}
