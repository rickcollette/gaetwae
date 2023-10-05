package algorithms

import (
    "github.com/rickcollette/gaetwae/shared"
)
var roundRobinIndex int

func RoundRobinBalancer() *shared.BackendInstance {
    index := roundRobinIndex % len(shared.GetBackendInstances())
    roundRobinIndex++
    return &shared.GetBackendInstances()[index]
}
