package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Servers interface {
	Address() string
	isAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

func (s *simpleServer) Address() string { return s.addr }

func (s *simpleServer) isAlive() bool { return true }

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func newSimpleServer(addrs string) *simpleServer {
	serverUrl, err := url.Parse(addrs)
	if err != nil {
		fmt.Printf("Error Occurred while parsing serverUrl from given address: %v\n", err)
		os.Exit(1)
	}

	return &simpleServer{
		addr:  addrs,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type loadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Servers
}

func newLoadBalancer(port string, servers []Servers) *loadBalancer {
	return &loadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func (lb *loadBalancer) getNextAvailableServer() Servers {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.isAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server

}

func (lb *loadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextAvailableServer()

	fmt.Printf("Forwarding Request to Address: %q\n", targetServer.Address())
	targetServer.Serve(rw, req)

}

func main() {
	servers := []Servers{
		newSimpleServer("file:///Users/arunepra/Desktop/Load-Balancer/HLD_LoadBalancer.html"),
		newSimpleServer("http://bing.com"),
		newSimpleServer("http://google.com"),
	}

	lb := newLoadBalancer("8000", servers)

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) { // '/' specifies the root dir.
		lb.serveProxy(rw, req) // lb forwards the request to the appropriate backend server and handles the proxy logic
	})

	fmt.Printf("Serving requests at 'localhost: %s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)

}
