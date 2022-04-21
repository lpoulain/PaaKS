package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

func index(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.URL.Scheme != "" {
		scheme = r.URL.Scheme
	}
	host, _, _ := net.SplitHostPort(r.Host)
	w.Header().Set("Access-Control-Allow-Origin", scheme+"://"+host+":3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	fmt.Println("URL:", r.URL.Path, "Method:", r.Method, "Scheme:", scheme)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	var req *http.Request
	var resp *http.Response
	var error error

	path := r.URL.Path

	if path != "/auth/login" {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(path, "/filesystem") && !strings.HasPrefix(path, "/svc-mgr") && !strings.HasPrefix(path, "/auth") {
			path = fmt.Sprintf("/tnt-%s-%s", auth[7:15], path[1:])
		}
	}

	switch r.Method {
	case "GET":
		req, error = http.NewRequest("GET", "http:/"+path, nil)
		//		resp, error = http.Get("http:/" + r.URL.Path)
		break
	case "POST":
		req, error = http.NewRequest("POST", "http:/"+path, r.Body)
		//		resp, error = http.Post("http:/"+r.URL.Path, r.Header.Get("Content-Type"), r.Body)
		break
	case "DELETE":
		req, error = http.NewRequest("DELETE", "http:/"+path, nil)
		break
	default:
		http.Error(w, "Unsupported method: "+r.Method, http.StatusBadRequest)
		return
	}

	if error != nil {
		http.Error(w, error.Error(), http.StatusBadRequest)
		return
	}

	if path != "/auth/login" {
		req.Header.Add("Authorization", r.Header.Get("Authorization"))
	}
	req.Header.Add("Content-Type", r.Header.Get("Content-Type"))

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, cookie := range resp.Cookies() {
		http.SetCookie(w, cookie)
	}

	fmt.Fprint(w, string(body))
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Health Check</h1>")
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/health_check", check)
	fmt.Println("Server starting...")
	http.ListenAndServe(":3000", nil)
}
