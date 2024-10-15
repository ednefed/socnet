package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var secret = []byte("secret")
var tokenLifetime int64
var feedSize int64
var feedUpdateBatchSize int64
var ctx = context.Background()
var dialogAPIHost string

var dsnHost = getEnvVar("POSTGRESQL_HOST", "localhost")
var dsnPort = getEnvVar("POSTGRESQL_PORT", "5432")
var dsnSlave1Host = getEnvVar("POSTGRESQL2_HOST", "localhost")
var dsnSlave1Port = getEnvVar("POSTGRESQL2_PORT", "5433")
var dsnSlave2Host = getEnvVar("POSTGRESQL3_HOST", "localhost")
var dsnSlave2Port = getEnvVar("POSTGRESQL3_PORT", "5434")
var dsnDb = getEnvVar("POSTGRESQL_DB", "postgres")
var dsnUsername = getEnvVar("POSTGRESQL_USERNAME", "postgres")
var dsnPassword = getEnvVar("POSTGRESQL_PASSWORD", "postgres")
var dsnSslMode = getEnvVar("POSTGRESQL_SSL_MODE", "disable")

func main() {
	db = connectToDB(dsnHost, dsnPort, dsnDb, dsnUsername, dsnPassword, dsnSslMode)
	db2 = connectToDB(dsnSlave1Host, dsnSlave1Port, dsnDb, dsnUsername, dsnPassword, dsnSslMode)
	db3 = connectToDB(dsnSlave2Host, dsnSlave2Port, dsnDb, dsnUsername, dsnPassword, dsnSslMode)
	dbMigrate()

	redisHost := getEnvVar("REDIS_HOST", "localhost")
	redisPort := getEnvVar("REDIS_PORT", "6379")
	redisPassword := getEnvVar("REDIS_PASSWORD", "")
	feedCache = connectToRedis(redisHost, redisPort, redisPassword)

	rabbitmqHost := getEnvVar("RABBITMQ_HOST", "localhost")
	rabbitmqPort := getEnvVar("RABBITMQ_PORT", "5672")
	rabbitmqUsername := getEnvVar("RABBITMQ_USERNAME", "admin")
	rabbitmqPassword := getEnvVar("RABBITMQ_PASSWORD", "admin")
	rabbitmq = connectToRabbitMQ(rabbitmqHost, rabbitmqPort, rabbitmqUsername, rabbitmqPassword)

	dialogAPIHost = getEnvVar("DIALOG_API_Host", "dialog_api:8080")

	var err error
	tokenLifetime, err = strconv.ParseInt(getEnvVar("TOKEN_LIFETIME", "60"), 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	feedSize, err = strconv.ParseInt(getEnvVar("FEED_SIZE", "10"), 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	feedUpdateBatchSize, err = strconv.ParseInt(getEnvVar("FEED_UPDATE_BATCH_SIZE", "10"), 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	go cacheSubPrepareFeedUpdate()
	go cacheSubFeedUpdate()

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
	router.PUT("/friend/:id", addFriendByID)
	router.DELETE("/friend/:id", deleteFriendByID)
	router.POST("/post", createPost)
	router.GET("/post/:id", getPostByID)
	router.PUT("/post/:id", updatePost)
	router.DELETE("/post/:id", deletePost)
	router.GET("/feed", getFeedForUser)
	router.POST("/feeds/reload", reloadFeeds)
	router.POST("/dialog/:id", dialogAPIProxyHandler)
	router.GET("/dialog/:id", dialogAPIProxyHandler)
	router.GET("/post/feed/:id", getNewPostsWS)
	router.GET("/", getHome)
	serverHost := getEnvVar("SERVER_HOST", "0.0.0.0")
	serverPort := getEnvVar("SERVER_PORT", "8080")
	server := fmt.Sprintf("%v:%v", serverHost, serverPort)
	router.Run(server)
}
