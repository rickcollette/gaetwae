package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rickcollette/gaetwae/pkg/algorithms"
	"github.com/rickcollette/gaetwae/pkg/cache"
	"github.com/rickcollette/gaetwae/pkg/ratelimit"
	"github.com/rickcollette/gaetwae/pkg/shared"
	"github.com/rickcollette/gaetwae/pkg/tls"
	"github.com/rickcollette/gaetwae/pkg/transform"
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
	Cache           CacheConfig `json:"cache"`
	Transformations struct {
		Request  transform.TransformationConfig `json:"request"`
		Response transform.TransformationConfig `json:"response"`
	} `json:"transformations"`
}

var (
	config   Config
	appCache cache.Cache // Changed the variable name to appCache
)

func subscribeToConfigUpdates(redisClient *redis.Client, channelName string) {
	pubsub := redisClient.Subscribe(context.Background(), channelName)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			fmt.Println("Error receiving message:", err)
			continue
		}

		err = updateConfigFromMessage(msg.Payload)
		if err != nil {
			fmt.Println("Error updating config:", err)
		}
	}
}

func updateConfigFromMessage(message string) error {
    var updatedConfig Config
    if err := json.Unmarshal([]byte(message), &updatedConfig); err != nil {
        return err
    }

	config = updatedConfig

    return nil
}

func main() {
	    // Retrieve Redis configuration from environment variables
		redisAddr := os.Getenv("REDIS_ADDR")
		redisPassword := os.Getenv("REDIS_PASSWORD")
		redisDB := os.Getenv("REDIS_DB")
	
		// Default values if environment variables are not set
		if redisAddr == "" {
			redisAddr = "localhost:6379" // Default Redis address
		}
		var redisDBNum int
		if redisDB == "" {
			redisDBNum = 0 // Default Redis DB number
		} else {
			redisDBNum, _ = strconv.Atoi(redisDB)
		}
	
		// Initialize Redis client
		redisClient := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDBNum,
		})

	go subscribeToConfigUpdates(redisClient, "config_channel")

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

func reverseProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.Cache.Enabled {
			if content, err := appCache.Get(r.URL.Path); err == nil {
				w.Write(content)
				return
			}
		}

		var backend *shared.BackendInstance // Declared the backend variable

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
		for _, header := range config.Transformations.Request.Headers {
			r.Header.Set(header.Key, header.Value)
		}

		// If using a reverse proxy, you can set up a ModifyResponse function to apply response transformations
		proxy.ModifyResponse = func(res *http.Response) error {
			for _, header := range config.Transformations.Response.Headers {
				res.Header.Set(header.Key, header.Value)
			}

			if config.Transformations.Response.Body.Type == "append" {
				body, _ := io.ReadAll(res.Body)
				body = append(body, config.Transformations.Response.Body.Content...)
				res.Body = io.NopCloser(bytes.NewReader(body))
			}

			// Handle other body transformation types as needed

			return nil
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
