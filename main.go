// Command http-anatomy serves the HTMX "HTTP anatomy" teaching app.
package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"http-anatomy/internal/store"
	"http-anatomy/internal/web"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := web.NewServer(store.New())

	log.Printf("http-anatomy listening on:")
	log.Printf("  local:   http://localhost:%s", port)
	if ip := lanIP(); ip != "" {
		log.Printf("  network: http://%s:%s", ip, port)
	}

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

// lanIP returns the host's primary non-loopback IPv4 address, or "" if none.
func lanIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return ""
}
