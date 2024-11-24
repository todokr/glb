package backend

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

// Target is a representation of a proxied server
type Target struct {
	URL             *url.URL
	healthCheckPath string
	healthy         bool
	mux             sync.RWMutex
	rp              *httputil.ReverseProxy
}

func (t *Target) Serve(w http.ResponseWriter, r *http.Request) {
	t.rp.ServeHTTP(w, r)
}

func (t *Target) HealthCheckPath() *url.URL {
	return t.URL.ResolveReference(&url.URL{Path: t.healthCheckPath})
}

func (t *Target) IsHealthy() bool {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return t.healthy
}

func (t *Target) SetHealthy(healthy bool) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.healthy = healthy
}

func NewTarget(targetUrl string, hcp string) *Target {
	u, err := url.Parse(targetUrl)
	if err != nil {
		log.Fatal(err)
	}
	return &Target{
		URL:             u,
		healthCheckPath: hcp,
		healthy:         true,
		rp:              httputil.NewSingleHostReverseProxy(u),
	}
}
