package paaks

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

const ROOT = "00000000-0000-0000-0000-000000000000"

var Roles = map[int]string{
	1: "USER",
	2: "ADMIN",
	3: "ROOT",
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
	Authorized bool   `json:authorized`
	Role       int    `json:"role"`
	Email      string `json:"email"`
	Tenant     string `json:"tenant"`
	Expiration int    `json:"exp"`
}

func IssueError(w http.ResponseWriter, message string, status int) {
	fmt.Println(message)
	http.Error(w, message, http.StatusBadRequest)
}

func GetSecretKey() string {
	return os.Getenv("SECRET_KEY")
}

func GetConnectionString() string {
	return os.Getenv("DB_CONNECTION_STRING")
}

func GetToken(r *http.Request) (*Token, error) {
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
			Email:      claims["email"].(string),
			Tenant:     claims["tenant"].(string),
			Expiration: int(claims["exp"].(float64)),
			Role:       int(claims["role"].(float64)),
		}

		return &paaksToken, nil
	} else {
		return nil, fmt.Errorf("Cannot get field out of the token")
	}
}