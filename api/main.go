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
	dbMigrate()

	citusHost := getEnvVar("CITUS_HOST", "localhost")
	citusPort := getEnvVar("CITUS_PORT", "5432")
	citusDb := getEnvVar("CITUS_DB", "postgres")
	citusUsername := getEnvVar("CITUS_USERNAME", "postgres")
	citusPassword := getEnvVar("CITUS_PASSWORD", "postgres")
	citusSslMode := getEnvVar("CITUS_SSL_MODE", "disable")
	citus = connectToDB(citusHost, citusPort, citusDb, citusUsername, citusPassword, citusSslMode)

	redisHost := getEnvVar("REDIS_HOST", "localhost")
	redisPort := getEnvVar("REDIS_PORT", "6379")
	redisPassword := getEnvVar("REDIS_PASSWORD", "")
	feedCache = connectToRedis(redisHost, redisPort, redisPassword)

	rabbitmqHost := getEnvVar("RABBITMQ_HOST", "localhost")
	rabbitmqPort := getEnvVar("RABBITMQ_PORT", "5672")
	rabbitmqUsername := getEnvVar("RABBITMQ_USERNAME", "admin")
	rabbitmqPassword := getEnvVar("RABBITMQ_PASSWORD", "admin")
	rabbitmq = connectToRabbitMQ(rabbitmqHost, rabbitmqPort, rabbitmqUsername, rabbitmqPassword)

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
	router.POST("/dialog/:id", createDialogMessage)
	router.GET("/dialog/:id", getDialogMessages)
	router.GET("/post/feed/:id", getNewPostsWS)
	router.GET("/", getHome)
	serverHost := getEnvVar("SERVER_HOST", "0.0.0.0")
	serverPort := getEnvVar("SERVER_PORT", "8080")
	server := fmt.Sprintf("%v:%v", serverHost, serverPort)
	router.Run(server)
}
