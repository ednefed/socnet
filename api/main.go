package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db, db2, db3 *sql.DB
var secret = []byte("secret")
var tokenLifetime int64

func main() {
	dsnHost := getEnvVar("POSTGRESQL_HOST", "localhost")
	dsnPort := getEnvVar("POSTGRESQL_PORT", "5432")
	dsnSlave1Host := getEnvVar("POSTGRESQL2_HOST", "localhost")
	dsnSlave1Port := getEnvVar("POSTGRESQL2_PORT", "5433")
	dsnSlave2Host := getEnvVar("POSTGRESQL3_HOST", "localhost")
	dsnSlave2Port := getEnvVar("POSTGRESQL3_PORT", "5434")
	dsnDb := getEnvVar("POSTGRESQL_DB", "postgres")
	dsnUsername := getEnvVar("POSTGRESQL_USERNAME", "postgres")
	dsnPassword := getEnvVar("POSTGRESQL_PASSWORD", "postgres")
	dsnSslMode := getEnvVar("POSTGRESQL_SSL_MODE", "disable")
	db = connectToDB(dsnHost, dsnPort, dsnDb, dsnUsername, dsnPassword, dsnSslMode)
	db2 = connectToDB(dsnSlave1Host, dsnSlave1Port, dsnDb, dsnUsername, dsnPassword, dsnSslMode)
	db3 = connectToDB(dsnSlave2Host, dsnSlave2Port, dsnDb, dsnUsername, dsnPassword, dsnSslMode)
	var err error
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
	router.GET("/user/search", searchUsersByFistAndLastName)
	serverHost := getEnvVar("SERVER_HOST", "0.0.0.0")
	serverPort := getEnvVar("SERVER_PORT", "8080")
	server := fmt.Sprintf("%v:%v", serverHost, serverPort)
	router.Run(server)
}
