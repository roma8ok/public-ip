package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type response struct {
	RemoteAddr    string `json:"remote_addr"`
	XForwardedFor string `json:"x_forwarded_for"`
	XRealIP       string `json:"x_real_ip"`
}

func handler(w http.ResponseWriter, req *http.Request) {
	res := response{
		RemoteAddr:    req.RemoteAddr,
		XForwardedFor: req.Header.Get("X-Forwarded-For"),
		XRealIP:       req.Header.Get("X-Real-Ip"),
	}

	resJSON, _ := json.Marshal(res)

	fmt.Fprintf(w, string(resJSON))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}
