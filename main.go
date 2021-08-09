package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// getIP tries to find the IP from ip:port string.
// If there is no IP in the string, getIP returns an empty string.
// getIP doesn't check the port for validity. This means getIP("127.0.0.1:100000") returns "127.0.0.1".
func getIP(s string) string {
	ip, _, err := net.SplitHostPort(s)
	if err != nil {
		ip2 := net.ParseIP(s)
		if ip2 == nil {
			return ""
		}
		return ip2.String()
	}

	ip3 := net.ParseIP(ip)
	if ip3 == nil {
		return ""
	}
	return ip3.String()
}

// getIPFromRequest returns the IP from header X-Forwarded-For or from header X-Real-Ip or from request.RemoteAddr.
// If there is no IP, getIPFromRequest returns an empty string.
func getIPFromRequest(req *http.Request) string {
	if xForwardedFor := req.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		if ip := getIP(xForwardedFor); ip != "" {
			return xForwardedFor
		}
	}

	if xRealIP := req.Header.Get("X-Real-Ip"); xRealIP != "" {
		if ip := getIP(xRealIP); ip != "" {
			return xRealIP
		}
	}

	return getIP(req.RemoteAddr)
}

func handler(w http.ResponseWriter, req *http.Request) {
	_, _ = io.WriteString(w, getIPFromRequest(req))
}

func main() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	logger := log.New(os.Stderr, "public-ip: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Print("Service started")

	http.HandleFunc("/", handler)

	go func() {
		if err := http.ListenAndServe(":80", nil); err != nil {
			logger.Print(err)
		}
	}()

	args := os.Args
	if len(args) == 3 {
		go func() {
			if err := http.ListenAndServeTLS(":443", args[1], args[2], nil); err != nil {
				logger.Print(err)
			}
		}()
	} else {
		logger.Print("Paths to certificate and private key files are required to start HTTPS. HTTPS is not started.")
	}

	if stop := <-exit; stop != nil {
		logger.Fatal("Service stopped")
	}
}
