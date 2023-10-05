package shared

type BackendInstance struct {
    URL       string
    Weight    int
    Connections int // Used for Least Connections
}
