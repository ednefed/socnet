package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

var secret = []byte("secret")

func main() {
	redisHost := getEnvVar("REDIS_HOST", "localhost")
	redisPort := getEnvVar("REDIS_PORT", "6379")
	redisPassword := getEnvVar("REDIS_PASSWORD", "")
	cache = connectToRedis(redisHost, redisPort, redisPassword)

	rabbitmqHost := getEnvVar("RABBITMQ_HOST", "localhost")
	rabbitmqPort := getEnvVar("RABBITMQ_PORT", "5672")
	rabbitmqUsername := getEnvVar("RABBITMQ_USERNAME", "admin")
	rabbitmqPassword := getEnvVar("RABBITMQ_PASSWORD", "admin")
	queue = connectToRabbitMQ(rabbitmqHost, rabbitmqPort, rabbitmqUsername, rabbitmqPassword)

	go queueSubscribeForOperations()

	if getEnvVar("ENVIRONMENT", "Development") == "Production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()
	router.GET("/api/v2/counter/:id", getUnreadMessagesCount)
	serverHost := getEnvVar("SERVER_HOST", "0.0.0.0")
	serverPort := getEnvVar("SERVER_PORT", "8080")
	server := fmt.Sprintf("%v:%v", serverHost, serverPort)
	router.Run(server)
}
