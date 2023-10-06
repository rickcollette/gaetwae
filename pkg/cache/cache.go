package cache

type Cache interface {
	Set(key string, value []byte, expirationTime int) error
	Get(key string) ([]byte, error)
}
