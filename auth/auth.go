package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"auth/paaks"
	"auth/paaksdb"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
)

const clientID = "paaks"

/*
func queryUsers(filters ...string) ([]User, error) {
	connStr := getConnectionString()
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("ERROR connecting to the database:", err.Error())
		return []User{}, err
	}

	sql := "SELECT email, password, salt, fullname, tenant FROM admin.users"
	args := []string{}
	if len(filters) > 0 {
		sql += " WHERE"
		for idx, filter := range filters {
			if idx%2 == 0 {
				sql += fmt.Sprintf(" AND %s = $%d", filter, idx/2+1)
			} else {
				args = append(args, filter)
			}
		}
	}

	stmt, err := conn.Prepare(sql)
	stmt.Query(args)

	rows, err := conn.Query(sql)

	if err != nil {
		fmt.Println("ERROR querying data:", err.Error())
		return []User{}, err
	}

	users := []User{}

	defer conn.Close()
	defer rows.Close()
	for rows.Next() {
		var email string
		var password string
		var salt string
		var fullname string
		var tenant string
		if err := rows.Scan(&email, &password, &salt, &fullname, &tenant); err != nil {
			return []User{}, err
		}

		users = append(users, User{
			email:        email,
			passwordhash: password,
			salt:         salt,
			fullname:     fullname,
			tenant:       tenant,
		})
	}

	return users, nil
}
*/

func userConstructor(rows *sql.Rows) (paaks.User, error) {
	var email string
	var password string
	var salt string
	var fullname string
	var tenant string
	var role int

	if err := rows.Scan(&email, &password, &salt, &fullname, &tenant, &role); err != nil {
		return paaks.User{}, err
	}

	return paaks.User{
		Email:        email,
		Passwordhash: password,
		Fullname:     fullname,
		Salt:         salt,
		Tenant:       tenant,
		Role:         role,
	}, nil
}

func GenerateJWT(user paaks.User) (string, error) {
	var mySigningKey = []byte(paaks.GetSecretKey())
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["email"] = user.Email
	claims["tenant"] = user.Tenant
	claims["role"] = user.Role
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
	token, err := paaks.GetToken(r)

	if err != nil {
		paaks.IssueError(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	tokenJson := map[string]interface{}{
		"email":      token.Email,
		"tenant":     token.Tenant,
		"role":       token.Role,
		"expiration": token.Expiration,
	}

	result, _ := json.Marshal(tokenJson)
	w.Write(result)
}

func login(w http.ResponseWriter, r *http.Request) {
	var authdetails Authentication
	err := json.NewDecoder(r.Body).Decode(&authdetails)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var users []paaks.User
	users, err = paaksdb.QueryDb[paaks.User]("SELECT email, password, salt, fullname, tenant, role FROM admin.users WHERE email = $1", userConstructor, authdetails.Email)

	if err != nil || len(users) != 1 {
		fmt.Println("Cannot find email '", authdetails.Email, "', ", len(users), "response")
		if err != nil {
			paaks.IssueError(w, err.Error(), http.StatusUnauthorized)
		} else {
			paaks.IssueError(w, "Invalid username or password", http.StatusUnauthorized)
		}
		return
	}

	user := users[0]
	hash := sha256.Sum256([]byte(user.Salt + authdetails.Password))
	computedPassword := base64.StdEncoding.EncodeToString(hash[:])

	if user.Passwordhash != computedPassword {
		paaks.IssueError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := GenerateJWT(user)
	if len(users) != 1 {
		paaks.IssueError(w, "Cannot generate token", http.StatusInternalServerError)
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
