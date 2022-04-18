package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func login(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := strings.TrimSpace(string(body))
	fmt.Printf("Auth for %s\n", token)
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
	http.SetCookie(w, &cookie)
	fmt.Fprint(w, string(token))
}

func logout(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(-1 * time.Hour)
	cookie := http.Cookie{Name: "token", Value: "", Expires: expiration}
	http.SetCookie(w, &cookie)
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Health Check</h1>")
}

func main() {
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/health_check", check)
	fmt.Println("Service starting...")
	http.ListenAndServe(":3000", nil)
}
