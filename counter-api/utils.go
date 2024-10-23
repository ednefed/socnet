package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func getEnvVar(key, fallback string) string {
	value := os.Getenv(key)

	if len(value) == 0 {
		return fallback
	}

	return value
}

func verifyToken(c *gin.Context) error {
	authorizationHeader := c.Request.Header["Authorization"]

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

	c.Set("user_id", claims["id"])
	return nil
}

func getUserIDFromContext(c *gin.Context) (int64, error) {
	userIDAny, exists := c.Get("user_id")

	if !exists {
		msg := "getUserIDFromContext: user_id not present in context"
		log.Println(msg)
		return 0, fmt.Errorf(msg)
	}

	userID, err := strconv.ParseInt(userIDAny.(string), 10, 64)

	if err != nil {
		msg := "getUserIDFromContext: user_id is not an int64"
		log.Println(msg)
		return 0, fmt.Errorf(msg)
	}

	return userID, nil
}
