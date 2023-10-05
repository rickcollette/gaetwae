package algorithms

import (
    "github.com/rickcollette/gaetwae/shared"
)

var weightedRoundRobinIndex int

func WeightedRoundRobinBalancer() *shared.BackendInstance {
    index := weightedRoundRobinIndex % len(backendInstances)
    weightedRoundRobinIndex++
    return &backendInstances[index]
}
