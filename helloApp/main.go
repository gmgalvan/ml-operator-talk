package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {

	server := http.NewServeMux()
	server.HandleFunc("/", hello)

	log.Printf("Server listening on port %v", 8080)
	err := http.ListenAndServe(":8080", server)
	log.Fatal(err)
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request: %s", r.URL.Path)
	host, _ := os.Hostname()
	fmt.Fprintf(w, "Hello World\n")
	fmt.Fprintf(w, "Version: 2.0.0\n")
	fmt.Fprintf(w, "Hostname: %s\n", host)

	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Fprintf(w, "IP Address: %s\n", ipnet.IP.String())
			}
		}
	}
}
