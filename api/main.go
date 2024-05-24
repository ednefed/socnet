package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB
var secret = []byte("secret")
var tokenLifetime int64

func main() {
	dsnHost := getEnvVar("POSTGRESQL_HOST", "localhost")
	dsnPort := getEnvVar("POSTGRESQL_PORT", "5432")
	dsnDb := getEnvVar("POSTGRESQL_DB", "postgres")
	dsnUsername := getEnvVar("POSTGRESQL_USERNAME", "postgres")
	dsnPassword := getEnvVar("POSTGRESQL_PASSWORD", "postgres")
	dsnSslMode := getEnvVar("POSTGRESQL_SSL_MODE", "disable")
	dsn := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v", dsnUsername, dsnPassword, dsnHost, dsnPort, dsnDb, dsnSslMode)
	var err error
	db, err = sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected to %v:%v/%v as %v", dsnHost, dsnPort, dsnDb, dsnUsername)
	tokenLifetime, err = strconv.ParseInt(getEnvVar("TOKEN_LIFETIME", "60"), 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	if getEnvVar("ENVIRONMENT", "Development") == "Production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()
	router.POST("/user", signup)
	router.GET("/user/:id", getUserByID)
	router.POST("/login", login)
	serverHost := getEnvVar("SERVER_HOST", "0.0.0.0")
	serverPort := getEnvVar("SERVER_PORT", "8080")
	server := fmt.Sprintf("%v:%v", serverHost, serverPort)
	router.Run(server)
}
