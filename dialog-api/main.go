package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

var secret = []byte("secret")

func main() {
	tarantoolHost := getEnvVar("TARANTOOL_HOST", "localhost")
	tarantoolPort := getEnvVar("TARANTOOL_PORT", "3301")
	tarantoolUsername := getEnvVar("TARANTOOL_USERNAME", "guest")
	tt = connectToTarantool(tarantoolHost, tarantoolPort, tarantoolUsername)

	if getEnvVar("ENVIRONMENT", "Development") == "Production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()
	router.POST("/api/v2/dialog/:id", createDialogMessage)
	router.GET("/api/v2/dialog/:id", getDialogMessages)
	serverHost := getEnvVar("SERVER_HOST", "0.0.0.0")
	serverPort := getEnvVar("SERVER_PORT", "8080")
	server := fmt.Sprintf("%v:%v", serverHost, serverPort)
	router.Run(server)
}
