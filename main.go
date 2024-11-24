package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"todokr.github.io/glb/backend"
	"todokr.github.io/glb/pool"
)

const HeathCheckInterval = 5 * time.Second
const HealthCheckTimeout = 1 * time.Second

func main() {
	var targets string
	var port int
	var healthCheckPath string
	flag.StringVar(&targets, "targets", "", "Target servers")
	flag.IntVar(&port, "port", 9090, "Port to serve")
	flag.StringVar(&healthCheckPath, "health-check-path", "/health", "Health check path")
	flag.Parse()

	if len(targets) == 0 {
		log.Fatal("no targets")
	}
	var urls []string
	for _, url := range strings.Split(targets, ",") {
		urls = append(urls, strings.TrimSpace(url))
	}

	fmt.Printf("Starting server on port %d\n", port)
	fmt.Printf("Targets: %v\n", urls)

	var ts []*backend.Target
	for _, url := range urls {
		ts = append(ts, backend.NewTarget(url, healthCheckPath))
	}

	pool := pool.NewRoundRobinServerPool(ts)
	healthCheck := func() {
		t := time.NewTicker(HeathCheckInterval)
		for {
			select {
			case <-t.C:
				log.Printf("Health check start ------------------------\n")
				for _, target := range ts {
					checkHealth(target)
				}
			}
		}
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		target := pool.Choose()
		if target == nil {
			http.Error(w, "no targets available", http.StatusServiceUnavailable)
			return
		}
		target.Serve(w, r)
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(handler),
	}
	go healthCheck()
	log.Fatal(server.ListenAndServe())
}

func checkHealth(target *backend.Target) {
	log.Printf("Checking %s\n", target.URL)
	c := http.Client{
		Timeout: time.Duration(HealthCheckTimeout),
	}
	res, err := c.Get(target.HealthCheckPath().String())

	if err != nil || res.StatusCode != http.StatusOK {
		body := res.Body
		defer body.Close()
		var msg string
		if m, err := io.ReadAll(body); err == nil {
			msg = string(m)
		} else {
			msg = err.Error()
		}

		log.Printf("[❌] %s code:%v, %v\n", target.URL, res.StatusCode, msg)
		target.SetHealthy(false)
	} else {
		target.SetHealthy(true)
		log.Printf("[✅] %s\n", target.URL)
	}
}
