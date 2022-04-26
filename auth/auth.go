package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
)

const clientID = "paaks"

func issueError(w http.ResponseWriter, message string) {
	fmt.Println(message)
	http.Error(w, message, http.StatusBadRequest)
}

func GenerateJWT(user User) (string, error) {
	var mySigningKey = []byte(getSecretKey())
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["email"] = user.email
	claims["tenant"] = user.tenant
	claims["role"] = user.role
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}
	return tokenString, nil
}

type Authentication struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func parseToken(w http.ResponseWriter, r *http.Request) {
	token, err := getToken(r)

	if err != nil {
		issueError(w, "Invalid token: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Email: %s, Tenant: %s, Role: %d, Expiration: %d", token.Email, token.Tenant, token.Role, token.Expiration)
}

func login(w http.ResponseWriter, r *http.Request) {
	var authdetails Authentication
	err := json.NewDecoder(r.Body).Decode(&authdetails)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	users, err := queryUsers("email = '" + authdetails.Email + "'")

	if err != nil || len(users) != 1 {
		fmt.Println("Cannot find email '", authdetails.Email, "', ", len(users), "response")
		if err != nil {
			issueError(w, err.Error())
		} else {
			issueError(w, "Invalid username or password")
		}
		return
	}

	user := users[0]
	hash := sha256.Sum256([]byte(user.salt + authdetails.Password))
	computedPassword := base64.StdEncoding.EncodeToString(hash[:])

	if user.passwordhash != computedPassword {
		issueError(w, "Invalid username or password")
		return
	}

	token, err := GenerateJWT(user)
	if len(users) != 1 {
		issueError(w, "Cannot generate token")
		return
	}

	fmt.Printf("Auth for %s\n", token)
	w.Header().Set("Content-Type", "text/plain")
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "token", Value: token, Expires: expiration}
	http.SetCookie(w, &cookie)

	fmt.Fprintln(w, token)
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
	http.HandleFunc("/token", parseToken)
	http.HandleFunc("/health_check", check)
	fmt.Println("Service starting...")
	http.ListenAndServe(":3000", nil)
}
