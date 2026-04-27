package specialists

import "sync"

// Default returns the process-wide default Registry. Specialists register
// themselves here in their init() functions; the orchestrator and MCP tool
// handlers read from here.
func Default() *Registry {
	defaultOnce.Do(func() {
		defaultRegistry = NewRegistry()
	})
	return defaultRegistry
}

var (
	defaultRegistry *Registry
	defaultOnce     sync.Once
)
