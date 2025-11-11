package testutils

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
)

func HTTPServer(routes map[string]http.HandlerFunc) (url.URL, func()) {
	mux := http.NewServeMux()

	for pattern, handler := range routes {
		mux.HandleFunc(pattern, handler)
	}

	server := &http.Server{
		Handler: mux,
	}

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	addr := listener.Addr().String()

	go func() {
		server.Serve(listener)
	}()

	parsedURL, _ := url.Parse(fmt.Sprintf("http://%s/", addr))

	return *parsedURL, func() {
		server.Close()
	}
}
