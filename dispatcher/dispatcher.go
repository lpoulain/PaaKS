package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/lpoulain/PaaKS/paaks"
)

var systemServices = map[string]bool{
	"filesystem": true,
	"svc-mgr":    true,
	"db-mgr":     true,
	"admin":      true,
	"auth":       true,
	"frontend":   true,
}

var unauthenticatedServices = map[string]bool{
	"auth":     true,
	"frontend": true,
}

func index(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.URL.Scheme != "" {
		scheme = r.URL.Scheme
	}
	host, _, _ := net.SplitHostPort(r.Host)

	referer := r.Header.Get("Referer")
	if referer != "" {
		hostStrings := strings.Split(referer, "/")
		if len(hostStrings) >= 3 {
			refererHost, refererPort, _ := net.SplitHostPort(hostStrings[2])
			if refererHost == host {
				w.Header().Set("Access-Control-Allow-Origin", scheme+"://"+refererHost+":"+refererPort)
			}
		}
	}

	//	w.Header().Set("Access-Control-Allow-Origin", "*")
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
	pathSubstrings := strings.Split(r.URL.Path, "/")
	if len(pathSubstrings) < 2 || pathSubstrings[1] == "" {
		paaks.IssueError(w, "Please specify a service", http.StatusBadRequest)
		return
	}
	service := pathSubstrings[1]

	token, err := paaks.GetToken(r)

	if _, ok := unauthenticatedServices[service]; !ok {
		if token == nil {
			paaks.IssueError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	if _, ok := systemServices[service]; !ok {
		path = fmt.Sprintf("/tnt-%s-%s", token.Tenant[:8], path[1:])
	}

	fmt.Println("  => http:/" + path)

	switch r.Method {
	case "GET", "DELETE":
		req, error = http.NewRequest(r.Method, "http:/"+path, nil)
		//		resp, error = http.Get("http:/" + r.URL.Path)
		break
	case "POST", "PUT":
		req, error = http.NewRequest(r.Method, "http:/"+path, r.Body)
		//		resp, error = http.Post("http:/"+r.URL.Path, r.Header.Get("Content-Type"), r.Body)
		break
	default:
		http.Error(w, "Unsupported method: "+r.Method, http.StatusBadRequest)
		return
	}

	if error != nil {
		paaks.IssueError(w, error.Error(), http.StatusBadRequest)
		return
	}

	if path != "/auth/login" {
		req.Header.Add("Authorization", getAuthorization(r))
	}
	req.Header.Add("Content-Type", r.Header.Get("Content-Type"))

	client := &http.Client{}
	resp, err = client.Do(req)

	if err != nil {
		fmt.Println("  => Error executing the request: " + err.Error())
		paaks.IssueError(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("  => Error reading the response: " + err.Error())
		paaks.IssueError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Body: ", body)

	for _, cookie := range resp.Cookies() {
		http.SetCookie(w, cookie)
	}

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Set(name, value)
		}
	}

	if resp.StatusCode != 200 {
		w.WriteHeader(resp.StatusCode)
	}
	w.Write(body)
}

func getAuthorization(r *http.Request) string {
	authString := r.Header.Get("Authorization")
	if authString == "" {
		for _, cookie := range r.Cookies() {
			if cookie.Name == "token" {
				authString = "Bearer " + strings.TrimSuffix(cookie.Value, "%0A")
			}
		}
	}

	return authString
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Health Check</h1>")
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/health_check", check)
	fmt.Println("Server starting...")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
