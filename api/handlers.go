package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func signup(c *gin.Context) {
	var user User

	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid data"})
		return
	}

	id, err := dbAddUser(user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func getUserByID(c *gin.Context) {
	if err := verifyToken(c.Request.Header["Authorization"]); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	var id int64
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
		return
	}

	user, err := dbGetUserByID(id)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
			return
		}
	}

	c.JSON(http.StatusOK, PrintableUser(user))
}

func login(c *gin.Context) {
	var userForAuth User

	if err := c.BindJSON(&userForAuth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid data"})
		return
	}

	userForCompare, err := dbGetUserByID(userForAuth.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	if userForCompare.ID != userForAuth.ID {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user or password"})
		return
	}

	hash := []byte(userForCompare.Password)
	password := []byte(userForAuth.Password)

	if err := bcrypt.CompareHashAndPassword(hash, password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Duration(tokenLifetime) * time.Minute).Unix(),
		"id":  userForCompare.ID,
	})

	tokenString, err := token.SignedString(secret)

	if err != nil {
		log.Printf("login.SignedString: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func searchUsersByFistAndLastName(c *gin.Context) {
	if err := verifyToken(c.Request.Header["Authorization"]); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	firstName := c.Query("first_name")
	lastName := c.Query("last_name")

	if len(firstName) == 0 || len(lastName) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameters"})
		return
	}

	users, err := dbGetUsersByFistAndLastName(firstName, lastName)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
			return
		}
	}

	c.JSON(http.StatusOK, users)
}
