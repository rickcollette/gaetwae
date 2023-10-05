package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
    "github.com/rickcollette/gaetwae/algorithms"
    "github.com/rickcollette/gaetwae/shared"
)


var backendInstances []shared.BackendInstance

func main() {
    // Read the configuration from the JSON file
    if err := loadConfig("gaetwae.conf"); err != nil {
        fmt.Println("Error loading configuration:", err)
        return
    }

    // Start the HTTP server on port 8080
    fmt.Println("Server is running on :8080...")
    http.HandleFunc("/", reverseProxyHandler())
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println(err)
    }
}

func reverseProxyHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Use a load balancing algorithm to select a backend instance
        backend := algorithms.WeightedRoundRobinBalancer() // You can choose the desired algorithm here
        backendURL := backend.URL

        // Create a reverse proxy for the selected backend service
        proxy := httputil.NewSingleHostReverseProxy(parseURL(backendURL))

        // Serve the request using the reverse proxy
        proxy.ServeHTTP(w, r)
    }
}

func loadConfig(filename string) error {
    // Read the JSON configuration file
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }

    // Unmarshal the JSON data into the backendInstances slice
    if err := json.Unmarshal(data, &backendInstances); err != nil {
        return err
    }

    return nil
}

func parseURL(rawURL string) *url.URL {
    // Parse the backend URL
    parsedURL, err := url.Parse(rawURL)
    if err != nil {
        panic(err)
    }
    return parsedURL
}
