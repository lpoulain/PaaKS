package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
)

func getSecretKey() string {
	return os.Getenv("SECRET_KEY")
}

func getToken(r *http.Request) (*Token, error) {
	var mySigningKey = []byte(getSecretKey())
	authorization := r.Header["Authorization"]
	if len(authorization) == 0 {
		return nil, fmt.Errorf("No authentication")
	}
	tokenString := authorization[0]
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return nil, fmt.Errorf("No authentication")
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	fmt.Println("Token: [" + tokenString + "]")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Cannot parse token")
		}

		return mySigningKey, nil
	})

	if err != nil {
		fmt.Println("Problem parsing token", err.Error())
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		paaksToken := Token{
			Email: claims["email"].(string),
			Tenant: claims["tenant"].(string),
			Expiration: int(claims["exp"].(float64)),
			Role: int(claims["role"].(float64)),
		}

		return &paaksToken, nil
	} else {
		return nil, fmt.Errorf("Cannot get field out of the token")
	}
}

type User struct {
	email        string
	passwordhash string
	salt         string
	fullname     string
	createDate   string
	tenant       string
	role         int
}

type Token struct {
	Authorized  bool `json:authorized`
	Role        int `json:"role"`
	Email       string `json:"email"`
	Tenant      string `json:"tenant"`
	Expiration  int `json:"exp"`
}

func getConnectionString() string {
	return os.Getenv("DB_CONNECTION_STRING")
}

func queryUsers(filter string) ([]User, error) {
	connStr := getConnectionString()
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("ERROR connecting to the database:", err.Error())
		return []User{}, err
	}

	sql := "SELECT email, password, salt, fullname, tenant FROM admin.users"
	if filter != "" {
		sql += " WHERE " + filter
	}

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
