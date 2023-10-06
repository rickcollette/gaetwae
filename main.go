package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rickcollette/gaetwae/pkg/algorithms"
	"github.com/rickcollette/gaetwae/pkg/cache"
	"github.com/rickcollette/gaetwae/pkg/ratelimit"
	"github.com/rickcollette/gaetwae/pkg/shared"
	"github.com/rickcollette/gaetwae/pkg/tls"
	"golang.org/x/time/rate"
)

type CacheConfig struct {
	Enabled        bool   `json:"enabled"`
	Type           string `json:"type"`
	ExpirationTime int    `json:"expirationTimeSeconds"`

	// For In-Memory Cache
	Capacity int `json:"capacity,omitempty"`

	// For Redis Cache
	Redis struct {
		Address  string `json:"address,omitempty"`
		Password string `json:"password,omitempty"`
		DB       int    `json:"db,omitempty"`
	} `json:"redis,omitempty"`

	// For Memcached Cache
	Memcached struct {
		Servers []string `json:"servers,omitempty"`
	} `json:"memcached,omitempty"`
}
type Config struct {
	LoadBalancingAlgorithm string                   `json:"loadBalancingAlgorithm"`
	Backends               []shared.BackendInstance `json:"backends"`
	TLS                    struct {
		CertPath string `json:"certPath"`
		KeyPath  string `json:"keyPath"`
	} `json:"tls"`
	Headers []struct {
		Name    string `json:"name"`
		Value   string `json:"value"`
		Enabled bool   `json:"enabled"`
	} `json:"headers"`
	RateLimiting struct {
		Enabled           bool `json:"enabled"`
		RequestsPerMinute int  `json:"requestsPerMinute"`
	} `json:"rateLimiting"`
	Cache CacheConfig `json:"cache"`
}

var (
    config Config
    appCache cache.Cache  // Changed the variable name to appCache
) 

func main() {
    err := loadConfig("gaetwae.conf")
    if err != nil {
        fmt.Println("Error loading configuration:", err)
        return
    }
if config.Cache.Enabled {
    switch config.Cache.Type {
    case "inMemory":
        appCache = cache.NewInMemoryCache(config.Cache.Capacity)
    case "redis":
        appCache = cache.NewRedisCache(config.Cache.Redis.Address, config.Cache.Redis.Password, config.Cache.Redis.DB)
    case "memcached":
        servers := strings.Join(config.Cache.Memcached.Servers, ",")
        appCache = cache.NewMemcachedCache(servers, strconv.Itoa(config.Cache.ExpirationTime))
    default:
        fmt.Println("Unsupported cache type")
        return
    }
}
	if config.RateLimiting.Enabled {
		r := rate.Every(1 * time.Minute / time.Duration(config.RateLimiting.RequestsPerMinute))
		limiter := rate.NewLimiter(r, int(config.RateLimiting.RequestsPerMinute))
		rateLimitMiddleware := ratelimit.RateLimitMiddleware(limiter)
		http.Handle("/", rateLimitMiddleware(reverseProxyHandler()))
	} else {
		http.HandleFunc("/", reverseProxyHandler())
	}
	if err := tls.StartHTTPSServer(nil, config.TLS.CertPath, config.TLS.KeyPath); err != nil {
		fmt.Println("Error starting HTTPS server:", err)
	}
}

func loadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	shared.SetBackendInstances(config.Backends)

	// Validate the load balancing algorithm
	validAlgorithms := []string{"leastConnections", "roundRobin", "weightedLeastConnections", "weightedRoundRobin"}
	valid := false
	for _, alg := range validAlgorithms {
		if config.LoadBalancingAlgorithm == alg {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid load balancing algorithm specified")
	}

	return nil
}

func reverseProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.Cache.Enabled {
            if content, err := appCache.Get(r.URL.Path); err == nil {
                w.Write(content)
                return
            }
		}

		var backend *shared.BackendInstance  // Declared the backend variable

		switch config.LoadBalancingAlgorithm {
		case "leastConnections":
			backend = algorithms.LeastConnectionsBalancer()
		case "roundRobin":
			backend = algorithms.RoundRobinBalancer()
		case "weightedLeastConnections":
			backend = algorithms.WeightedLeastConnectionsBalancer()
		case "weightedRoundRobin":
			backend = algorithms.WeightedRoundRobinBalancer()
		}

		backendURL := backend.URL
		proxy := httputil.NewSingleHostReverseProxy(parseURL(backendURL))

		for _, header := range config.Headers {
			if header.Enabled {
				r.Header.Set(header.Name, header.Value)
			}
		}

		proxy.ServeHTTP(w, r)
	}
}

func parseURL(rawURL string) *url.URL {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return parsedURL
}
