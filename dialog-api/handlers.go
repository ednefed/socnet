package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func createDialogMessage(c *gin.Context) {
	xRequestID := c.GetHeader("X-Request-ID")

	if err := verifyToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		log.Printf("request-id: %s, createDialogMessage.verifyToken: %v", xRequestID, err)
		return
	}

	userID, err := getUserIDFromContext(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		log.Printf("request-id: %s, createDialogMessage.getUserIDFromContext: %v", xRequestID, err)
		return
	}

	friendID, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
		log.Printf("request-id: %s, createDialogMessage.strconv.ParseInt: %v", xRequestID, err)
		return
	}

	var message Message

	if err := c.BindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid data"})
		log.Printf("request-id: %s, createDialogMessage.BindJSON: %v", xRequestID, err)
		return
	}

	message.FromUserID = userID
	message.ToUserID = friendID
	message.CreatedAt = time.Now().Format(time.RFC3339)

	if _, err := tarantoolCreateMessage(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		log.Printf("request-id: %s, createDialogMessage.tarantoolCreateMessage: %v", xRequestID, err)
		return
	}

	log.Printf("request-id: %s, createDialogMessage: Message sent", xRequestID)
	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

func getDialogMessages(c *gin.Context) {
	xRequestID := c.GetHeader("X-Request-ID")

	if err := verifyToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		log.Printf("request-id: %s, getDialogMessages.verifyToken: %v", xRequestID, err)
		return
	}

	userID, err := getUserIDFromContext(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		log.Printf("request-id: %s, getDialogMessages.getUserIDFromContext: %v", xRequestID, err)
		return
	}

	friendID, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
		log.Printf("request-id: %s, getDialogMessages.strconv.ParseInt: %v", xRequestID, err)
		return
	}

	messages, err := tarantoolGetDialogMessages(userID, friendID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		log.Printf("request-id: %s, getDialogMessages.tarantoolGetDialogMessages: %v", xRequestID, err)
		return
	}

	log.Printf("request-id: %s, getDialogMessages: Messages received", xRequestID)
	c.JSON(http.StatusOK, messages)
}
