package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// allowedTarget is the only CONNECT destination this proxy will ever relay to.
// It is derived from the same KUBERNETES_SERVICE_HOST/PORT env vars every pod
// gets auto-injected with, so the allowlist always matches the real apiserver
// wherever this is deployed - nothing is hardcoded.
func allowedTarget() string {
	return net.JoinHostPort(os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT"))
}

func connectHandler(allowed string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect || r.Host != allowed {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		dst, err := net.DialTimeout("tcp", r.Host, 5*time.Second)
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		defer dst.Close()

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijack unsupported", http.StatusInternalServerError)
			return
		}
		src, _, err := hijacker.Hijack()
		if err != nil {
			return
		}
		defer src.Close()

		src.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

		done := make(chan struct{}, 2)
		go func() { io.Copy(dst, src); done <- struct{}{} }()
		go func() { io.Copy(src, dst); done <- struct{}{} }()
		<-done
	}
}

func main() {
	allowed := allowedTarget()
	log.Printf("kube-apiserver-connect-proxy listening on :8080, allowed CONNECT target: %s", allowed)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           connectHandler(allowed),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
