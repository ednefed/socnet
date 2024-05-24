package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
)

func getEnvVar(key, fallback string) string {
	value := os.Getenv(key)

	if len(value) == 0 {
		return fallback
	}

	return value
}

func verifyToken(authorizationHeader []string) error {
	if authorizationHeader == nil {
		return fmt.Errorf("verifyToken: Authorization header not present")
	}

	bearer, tokenString := strings.Split(authorizationHeader[0], " ")[0], strings.Split(authorizationHeader[0], " ")[1]

	if bearer != "Bearer" {
		return fmt.Errorf("verifyToken: Unknown authorization header")
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return secret, nil
	})

	if err != nil {
		return err
	}

	return nil
}
