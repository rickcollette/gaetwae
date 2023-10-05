package shared

var backendInstances []BackendInstance

func GetBackendInstances() []BackendInstance {
    return backendInstances
}

func SetBackendInstances(instances []BackendInstance) {
    backendInstances = instances
}
