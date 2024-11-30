package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"todokr.github.io/glb/backend"
	"todokr.github.io/glb/pool"
)

const HeathCheckInterval = 1 * time.Second
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

	var ts []*backend.Target
	for _, url := range urls {
		ts = append(ts, backend.NewTarget(url, healthCheckPath))
	}

	printInfo := func() {
		fmt.Print("\033[H\033[J")
		fmt.Printf("Load balancer runs on port %d\n", port)
		fmt.Printf("Targets: %v\n", urls)
	}
	check := func(target *backend.Target) {
		c := http.Client{
			Timeout: time.Duration(HealthCheckTimeout),
		}
		res, err := c.Get(target.HealthCheckPath().String())

		if err != nil || res.StatusCode != http.StatusOK {
			body := res.Body
			defer body.Close()
			fmt.Printf("[❌] %s (%d)\n", target.URL, res.StatusCode)
			target.SetHealthy(false)
		} else {
			target.SetHealthy(true)
			fmt.Printf("[✅] %s\n", target.URL)
		}
	}

	pool := pool.NewRoundRobinServerPool(ts)
	healthCheck := func() {
		t := time.NewTicker(HeathCheckInterval)
		for {
			select {
			case <-t.C:
				printInfo()
				for _, target := range ts {
					check(target)
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
	printInfo()
	go healthCheck()
	log.Fatal(server.ListenAndServe())
}
