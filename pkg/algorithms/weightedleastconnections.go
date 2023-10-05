package algorithms
import (
	"github.com/rickcollette/gaetwae/shared"
)

func WeightedLeastConnectionsBalancer() *shared.BackendInstance {
    var selected *shared.BackendInstance
    minWeightedConnections := -1

    for _, backend := range shared.GetBackendInstances() {
        if selected == nil || (backend.Weight * backend.Connections) < minWeightedConnections {
            selected = &backend
            minWeightedConnections = backend.Weight * backend.Connections
        }
    }

    // Increment the connections count for the selected backend
    selected.Connections++
    return selected
}
