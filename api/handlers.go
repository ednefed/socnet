package main

import (
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
		log.Printf("signUp.BindJSON: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id, err := dbAddUser(user)

	if err != nil {
		log.Printf("signUp.dbRegisterUser: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func getUserByID(c *gin.Context) {
	if err := verifyToken(c.Request.Header["Authorization"]); err != nil {
		log.Printf("getUserByID.verifyToken: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var id int64
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		log.Printf("getUserByID.ParseInt: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	user, err := dbGetUserByID(id)

	if err != nil {
		log.Printf("getUserByID.dbGetUserByID: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if user.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else {
		c.JSON(http.StatusOK, PrintableUser(user))
	}
}

func login(c *gin.Context) {
	var userForAuth User

	if err := c.BindJSON(&userForAuth); err != nil {
		log.Printf("login.BindJSON: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	userForCompare, err := dbGetUserByID(userForAuth.ID)

	if err != nil {
		log.Printf("login.dbGetUserByID: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if userForCompare.ID != userForAuth.ID {
		log.Printf("login: user '%v' not found", userForAuth.ID)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	hash := []byte(userForCompare.Password)
	password := []byte(userForAuth.Password)

	if err := bcrypt.CompareHashAndPassword(hash, password); err != nil {
		log.Printf("login.CompareHashAndPassword: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Duration(tokenLifetime) * time.Minute).Unix(),
		"id":  userForCompare.ID,
	})

	tokenString, err := token.SignedString(secret)

	if err != nil {
		log.Printf("login.SignedString: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
