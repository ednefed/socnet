package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
)

var secret = []byte("secret")

func main() {
	tarantoolHost := getEnvVar("TARANTOOL_HOST", "localhost")
	tarantoolPort := getEnvVar("TARANTOOL_PORT", "3301")
	tarantoolUsername := getEnvVar("TARANTOOL_USERNAME", "guest")
	tt = connectToTarantool(tarantoolHost, tarantoolPort, tarantoolUsername)

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
	metrics := ginmetrics.GetMonitor()
	metrics.SetMetricPath("/metrics")
	metrics.SetDuration([]float64{0.01, 0.1, 0.3, 1.2, 5, 10})
	metrics.Use(router)
	router.POST("/api/v2/dialog/:id", createDialogMessage)
	router.GET("/api/v2/dialog/:id", getDialogMessages)
	serverHost := getEnvVar("SERVER_HOST", "0.0.0.0")
	serverPort := getEnvVar("SERVER_PORT", "8080")
	server := fmt.Sprintf("%v:%v", serverHost, serverPort)
	router.Run(server)
}
