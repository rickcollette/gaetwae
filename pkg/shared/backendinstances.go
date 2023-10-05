
package shared

// BackendInstance represents a single backend instance
type BackendInstance struct {
    Name        string
    URL         string
    Connections int
    Weight      int
}
// backendInstances holds the current backend instances
var backendInstances []BackendInstance

// GetBackendInstances returns the current backend instances
func GetBackendInstances() []BackendInstance {
    return backendInstances
}

// SetBackendInstances sets the current backend instances
func SetBackendInstances(instances []BackendInstance) {
    backendInstances = instances
}
