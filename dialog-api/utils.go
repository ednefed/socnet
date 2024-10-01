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

func convertTarantoolIntToInt64(data interface{}) (int64, error) {
	var result int64

	switch t := data.(type) {
	case uint:
		result = int64(data.(uint))
	case uint8:
		result = int64(data.(uint8))
	case uint16:
		result = int64(data.(uint16))
	case uint32:
		result = int64(data.(uint32))
	case uint64:
		result = int64(data.(uint64))
	case int:
		result = int64(data.(int))
	case int8:
		result = int64(data.(int8))
	case int16:
		result = int64(data.(int16))
	case int32:
		result = int64(data.(int32))
	case int64:
		result = int64(data.(int64))
	default:
		return 0, fmt.Errorf("convertIntToInt64: unsupported type %T", t)
	}

	return result, nil
}
