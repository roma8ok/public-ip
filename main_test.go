package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testRemoteAddrIPPort     = "10.10.10.10:10000"
	testRemoteAddrIP         = "10.10.10.10"
	testXForwardedForCorrect = "100.100.100.100"
	testXForwardedForWrong   = "wrong_x_forwarded_for"
	testXRealIPCorrect       = "200.200.200.200"
	testXRealIPWrong         = "wrong_x_real_ip"
)

func TestParseIP_WrongIPs(t *testing.T) {
	testCases := []string{
		"wrong_ip",
		"121.11.11.11:port:port",
		"121.11.11.11:777:777",
		"ip:7777",
		"212121.11.11.11",
		"212121.11.11.11:777",
	}

	for _, testCase := range testCases {
		if ip := getIP(testCase); ip != "" {
			t.Errorf(`getIP(%s) must returns "", but returns %s`, testCase, ip)
		}
	}
}

func TestParseIP_RightIPs(t *testing.T) {
	testCases := [][]string{
		{"127.0.0.1", "127.0.0.1"},
		{"127.0.0.1:1100", "127.0.0.1"},
		{"127.0.0.1:100000", "127.0.0.1"},
		{"182.212.0.1", "182.212.0.1"},
		{"182.212.0.1:7253", "182.212.0.1"},
		{"123.0.0.200:70000", "123.0.0.200"}, // getIP don't check port for validity
	}

	for idx := range testCases {
		if ip := getIP(testCases[idx][0]); ip != testCases[idx][1] {
			t.Errorf(`getIP(%s) must returns "%s", but returns "%s"`, testCases[idx][0], testCases[idx][1], ip)
		}
	}
}

func TestGetIPFromRequest_RightHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	req.RemoteAddr = testRemoteAddrIPPort
	if ip := getIPFromRequest(req); ip != testRemoteAddrIP {
		t.Errorf(`getIPFromRequest without headers X-Forwarded-For and X-Real-Ip must return IP from req.RemoteAddr, but returns "%s"`, ip)
	}

	req.Header.Set("X-Real-Ip", testXRealIPCorrect)
	if ip := getIPFromRequest(req); ip != testXRealIPCorrect {
		t.Errorf(`getIPFromRequest within header X-Real-Ip and without header X-Forwarded-For must return header X-Real-Ip, but returns "%s"`, ip)
	}

	req.Header.Set("X-Forwarded-For", testXForwardedForCorrect)
	if ip := getIPFromRequest(req); ip != testXForwardedForCorrect {
		t.Errorf(`getIPFromRequest within header X-Forwarded-For must return header X-Forwarded-For, but returns "%s"`, ip)
	}
}

func TestGetIPFromRequest_WrongHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	req.RemoteAddr = testRemoteAddrIPPort
	if ip := getIPFromRequest(req); ip != testRemoteAddrIP {
		t.Errorf(`getIPFromRequest without headers X-Forwarded-For and X-Real-Ip must return IP from req.RemoteAddr, but returns "%s"`, ip)
	}

	req.Header.Set("X-Real-Ip", testXRealIPWrong)
	if ip := getIPFromRequest(req); ip != testRemoteAddrIP {
		t.Errorf(`getIPFromRequest without correct X-Forwarded-For and X-Real-Ip must return IP from req.RemoteAddr, but returns "%s"`, ip)
	}

	req.Header.Set("X-Forwarded-For", testXForwardedForWrong)
	if ip := getIPFromRequest(req); ip != testRemoteAddrIP {
		t.Errorf(`getIPFromRequest without correct X-Forwarded-For and X-Real-Ip must return IP from req.RemoteAddr, but returns "%s"`, ip)
	}
}

func TestHandler_RemoteAddr(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	req.RemoteAddr = testRemoteAddrIPPort

	handler(res, req)

	result := res.Result()

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Errorf("can't read body from handler(). err = %s", err)
	}
	_ = result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf("handler() must returns status code 200, but returns %d", result.StatusCode)
	}
	if string(body) != testRemoteAddrIP {
		t.Errorf(`handler() with request without headers X-Forwarded-For and X-Real-Ip must return IP from req.RemoteAddr, but returns "%s"`, string(body))
	}
}

func TestHandler_CorrectXRealIP(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	req.RemoteAddr = testRemoteAddrIPPort
	req.Header.Set("X-Real-Ip", testXRealIPCorrect)

	handler(res, req)

	result := res.Result()

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Errorf("can't read body from handler(). err = %s", err)
	}
	_ = result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf("handler() must returns status code 200, but returns %d", result.StatusCode)
	}
	if string(body) != testXRealIPCorrect {
		t.Errorf(`handler() with request within header X-Real-Ip and without header X-Forwarded-For must returns IP from header X-Real-Ip, but returns "%s"`, string(body))
	}
}

func TestHandler_WrongXRealIP(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	req.RemoteAddr = testRemoteAddrIPPort
	req.Header.Set("X-Real-Ip", testXRealIPWrong)

	handler(res, req)

	result := res.Result()

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Errorf("can't read body from handler(). err = %s", err)
	}
	_ = result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf("handler() must returns status code 200, but returns %d", result.StatusCode)
	}
	if string(body) != testRemoteAddrIP {
		t.Errorf(`handler() with request without correct headers X-Forwarded-For and X-Real-Ip must return IP from req.RemoteAddr, but returns "%s"`, string(body))
	}
}

func TestHandler_CorrectXForwarderFor(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	req.RemoteAddr = testRemoteAddrIPPort
	req.Header.Set("X-Forwarded-For", testXForwardedForCorrect)

	handler(res, req)

	result := res.Result()

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Errorf("can't read body from handler(). err = %s", err)
	}
	_ = result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf("handler() must returns status code 200, but returns %d", result.StatusCode)
	}
	if string(body) != testXForwardedForCorrect {
		t.Errorf(`handler() with request within header X-Forwarded-For must returns IP from header X-Forwarded-For, but returns "%s"`, string(body))
	}
}

func TestHandler_WrongXForwarderFor(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	req.RemoteAddr = testRemoteAddrIPPort
	req.Header.Set("X-Real-Ip", testXRealIPCorrect)
	req.Header.Set("X-Forwarded-For", testXForwardedForWrong)

	handler(res, req)

	result := res.Result()

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Errorf("can't read body from handler(). err = %s", err)
	}
	_ = result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf("handler() must returns status code 200, but returns %d", result.StatusCode)
	}
	if string(body) != testXRealIPCorrect {
		t.Errorf(`handler() with request within correct header X-Real-Ip and without correct header X-Forwarded-For must return IP from header X-Real-Ip, but returns "%s"`, string(body))
	}
}
