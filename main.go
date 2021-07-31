package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

func parseIP(s string) string {
	ip, _, err := net.SplitHostPort(s)
	if err == nil {
		return ip
	}

	ip2 := net.ParseIP(s)
	if ip2 == nil {
		return ""
	}
	return ip2.String()
}

type response struct {
	RemoteAddr    string `json:"remote_addr"`
	XForwardedFor string `json:"x_forwarded_for"`
	XRealIP       string `json:"x_real_ip"`
}

func handler(w http.ResponseWriter, req *http.Request) {
	res := response{
		RemoteAddr:    parseIP(req.RemoteAddr),
		XForwardedFor: parseIP(req.Header.Get("X-Forwarded-For")),
		XRealIP:       parseIP(req.Header.Get("X-Real-Ip")),
	}

	resJSON, _ := json.Marshal(res)

	fmt.Fprintf(w, string(resJSON))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}
