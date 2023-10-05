package algorithms

import (
    "github.com/rickcollette/gaetwae/shared"
)

var weightedRoundRobinIndex int

func WeightedRoundRobinBalancer() *shared.BackendInstance {
    index := weightedRoundRobinIndex % len(shared.GetBackendInstances())
    weightedRoundRobinIndex++
    return &shared.GetBackendInstances()[index]
}
