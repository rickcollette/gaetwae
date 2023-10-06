package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/rickcollette/gaetwae/pkg/algorithms"
	"github.com/rickcollette/gaetwae/pkg/shared"
	"github.com/rickcollette/gaetwae/pkg/tls"
)

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
}

var config Config

func main() {
	if err := loadConfig("gaetwae.conf"); err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	fmt.Println("Server is running on :8080...")
	http.HandleFunc("/", reverseProxyHandler())

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
		var backend *shared.BackendInstance

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
