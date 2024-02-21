package jwt

import (
	"fmt"
	"log"
	"os"

	"github.com/go-chi/jwtauth/v5"
)

func Init() (*jwtauth.JWTAuth, error) {
	secret, err := os.ReadFile("C:/Users/Leonid/Desktop/portal/internal/lib/jwt/secret.txt") // Пока локальное положение ключа
	if err != nil {
		log.Println(err)
	}
	tokenAuth := jwtauth.New("HS256", secret, nil)
	return tokenAuth, nil
}

func New(tokenAuth *jwtauth.JWTAuth) (string, error) {

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123}) // Пример
	fmt.Println(tokenString)
	return tokenString, nil
}
