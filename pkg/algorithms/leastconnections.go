package algorithms

import (
	"github.com/rickcollette/gaetwae/pkg/shared"
)

func LeastConnectionsBalancer() *shared.BackendInstance {
	var selected *shared.BackendInstance
	minConnections := -1

	for _, backend := range shared.GetBackendInstances() {
		if selected == nil || backend.Connections < minConnections {
			selected = &backend
			minConnections = backend.Connections
		}
	}

	// Increment the connections count for the selected backend
	selected.Connections++
	return selected
}
