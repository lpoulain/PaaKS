package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/lpoulain/PaaKS/paaks"
	"github.com/lpoulain/PaaKS/paaksdb"
)

type Tenant struct {
	id   string
	name string
}

func tenantConstructor(rows *sql.Rows) (interface{}, error) {
	var id string
	var name string

	if err := rows.Scan(&id, &name); err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"id":   id,
		"name": name,
	}, nil
}

func userConstructor(rows *sql.Rows) (interface{}, error) {
	var email string
	var fullname string
	var tenant string
	var role int

	if err := rows.Scan(&email, &fullname, &tenant, &role); err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"email":    email,
		"fullname": fullname,
		"tenant":   tenant,
		"role":     paaks.Roles[role],
	}, nil
}

////////////////////////////////////////////////////
// Creating users
////////////////////////////////////////////////////

type NewUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Fullname string `json:"fullname"`
	Tenant   string `json:"tenant"`
}

func createUser(user NewUser) error {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 20)
	rand.Read(b)
	salt := fmt.Sprintf("%x", b)[:20]

	hash := sha256.Sum256([]byte(salt + user.Password))
	storedPassword := base64.StdEncoding.EncodeToString(hash[:])

	fmt.Println("Pw to store: '" + salt + user.Password + "'")
	fmt.Println("Hash:", hash)

	return paaksdb.ExecDb("INSERT INTO admin.users (email, password, salt, fullname, tenant, role) VALUES ($1, $2, $3, $4, $5, 1)", user.Email, storedPassword, salt, user.Fullname, user.Tenant)
}

////////////////////////////////////////////////////
// Handlers
////////////////////////////////////////////////////

func tenants(w http.ResponseWriter, r *http.Request) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if token.Tenant != paaks.Root {
		paaks.IssueError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	paaksdb.QueryDbToResponse(w, "SELECT id, name FROM admin.tenants", tenantConstructor)
}

func users(w http.ResponseWriter, r *http.Request) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if token.Tenant == paaks.Root {
		paaksdb.QueryDbToResponse(w, "SELECT email, fullname, tenant, role FROM admin.users", userConstructor)
	} else {
		paaksdb.QueryDbToResponse(w, "SELECT email, fullname, tenant, role FROM admin.users WHERE tenant = $1", userConstructor, token.Tenant)
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	var user NewUser
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		paaks.IssueError(w, "Cannot decode message", http.StatusBadRequest)
		return
	}

	err = createUser(user)

	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "User created\n")
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Health Check</h1>")
}

////////////////////////////////////////////////////

func main() {
	http.HandleFunc("/tenants", tenants)
	http.HandleFunc("/users", users)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/health_check", check)
	fmt.Println("Service starting...")
	http.ListenAndServe(":3000", nil)
}
