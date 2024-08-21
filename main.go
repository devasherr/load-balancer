package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	Address    string
	ActiveConn int
}

type loadBalancer struct {
	Backends []*Backend
	mutex    sync.Mutex
}

func (lb *loadBalancer) GetLeastConnBackend() *Backend {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	var selected *Backend
	minConn := int(^uint(0) >> 1)

	for _, backend := range lb.Backends {
		if backend.ActiveConn < minConn {
			minConn = backend.ActiveConn
			selected = backend
		}
	}

	return selected
}

func (lb *loadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.GetLeastConnBackend()

	// increase the connection count
	backend.ActiveConn++
	defer func() {
		// decrease the connection count after request is handled
		backend.ActiveConn--
	}()

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   backend.Address,
	})

	proxy.ServeHTTP(w, r)
}

func main() {
	backends := []*Backend{
		{Address: "http://localhost:3031", ActiveConn: 0},
		{Address: "http://localhost:3032", ActiveConn: 0},
		{Address: "http://localhost:3033", ActiveConn: 0},
	}

	lb := &loadBalancer{Backends: backends}
	fmt.Println("load balancer started on localhost:8800")
	http.ListenAndServe(":8080", lb)
}
