package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getUnreadMessagesCount(c *gin.Context) {
	if err := verifyToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	userID, err := getUserIDFromContext(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	friendID, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
		return
	}

	count, err := cacheGetUnreadMessagesCount(friendID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	dialog := Dialog{FromUserID: friendID, ToUserID: userID, Unread: count}

	c.JSON(http.StatusOK, dialog)
}
