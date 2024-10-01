package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

var upgrader = websocket.Upgrader{}

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
	if err := verifyToken(c); err != nil {
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
		"id":  strconv.FormatInt(userForCompare.ID, 10),
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
	if err := verifyToken(c); err != nil {
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

func addFriendByID(c *gin.Context) {
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

	if err := dbAddFriendByID(userID, friendID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Friend added"})
}

func deleteFriendByID(c *gin.Context) {
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

	if err := dbDeleteFriendByID(userID, friendID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend deleted"})
}

func createPost(c *gin.Context) {
	if err := verifyToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	userID, err := getUserIDFromContext(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	var post Post

	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid data"})
		return
	}

	post, err = dbCreatePost(post, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	c.JSON(http.StatusCreated, post)

	initiator := FeedUpdateInitiator{
		PostID:   post.ID,
		FriendID: post.UserID,
	}

	if err = cachePubPrepareFeedUpdate(initiator); err != nil {
		log.Printf("createPost.cachePubPrepareFeedUpdate: %v", err)
	}

	if err = rmqPubPost(post); err != nil {
		log.Printf("createPost.rmqPubPost: %v", err)
	}
}

func getPostByID(c *gin.Context) {
	if err := verifyToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
		return
	}

	post, err := dbGetPostByID(postID)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
			return
		}
	}

	c.JSON(http.StatusOK, post)
}

func updatePost(c *gin.Context) {
	if err := verifyToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	userID, err := getUserIDFromContext(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
		return
	}

	var post, updatedPost Post

	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid data"})
		return
	}

	post.ID = postID
	post.UserID = userID

	if updatedPost, err = dbUpdatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	if updatedPost.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func deletePost(c *gin.Context) {
	if err := verifyToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	userID, err := getUserIDFromContext(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
		return
	}

	var post = Post{ID: postID, UserID: userID}
	var deletedPost Post

	if deletedPost, err = dbDeletePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	if deletedPost.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}

func getFeedForUser(c *gin.Context) {
	if err := verifyToken(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	userID, err := getUserIDFromContext(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	var offset, limit int64
	offsetStr := c.Query("offset")

	if offsetStr == "" {
		offset = 0
	} else {
		offset, err = strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
			return
		}
	}

	limitStr := c.Query("limit")

	if limitStr == "" {
		limit = 5
	} else {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
			return
		}
	}

	posts, err := cacheGetFeedForUser(userID, offset, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

func reloadFeeds(c *gin.Context) {
	if err := feedCache.FlushDB(ctx).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	userIDs, err := dbGetUsersWithFriends()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	for _, userID := range userIDs {
		posts, err := dbGetPostsForFeedByUserID(userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
			return
		}

		for _, post := range posts {
			payload, err := json.Marshal(post)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
				return
			}

			if err := feedCache.RPush(ctx, strconv.FormatInt(userID, 10), payload).Err(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feeds reloaded"})
}

func dialogAPIProxyHandler(c *gin.Context) {
	remote, err := url.Parse(fmt.Sprintf("http://%s/api/v2/%s", dialogAPIHost, c.Request.URL.Path[1:]))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	xRequestID := c.GetHeader("X-Request-ID")

	if len(xRequestID) == 0 {
		xRequestID = uuid.New().String()
	}

	log.Printf("request-id: %s, proxy: %s %s -> %s", xRequestID, c.Request.Method, c.Request.URL.Path, dialogAPIHost)
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Header["X-Request-ID"] = []string{xRequestID}
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = remote.Path
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func getNewPostsWS(c *gin.Context) {
	// TODO: authentication on home page
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid parameter"})
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	defer ws.Close()

	friends, err := dbGetFriendsByUserID(userID)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("getNewPostsWS.dbGetFriendsByUserID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
			return
		}
	}

	if len(friends) == 0 {
		return
	}

	channel, err := rabbitmq.Channel()

	if err != nil {
		log.Printf("getNewPostsWS.Channel: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	defer channel.Close()

	queue_name := strconv.FormatInt(userID, 10) + "-" + strconv.Itoa(rand.Int())[0:8]
	queue, err := channel.QueueDeclare(
		queue_name,
		false,
		true,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Printf("getNewPostsWS.QueueDeclare: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	for _, friend := range friends {
		err = channel.QueueBind(
			queue.Name,
			strconv.FormatInt(friend, 10),
			"amq.direct",
			false,
			nil,
		)

		if err != nil {
			log.Printf("rmqSubPosts.QueueBind: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
			return
		}
	}

	msgs, err := channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Printf("getNewPostsWS.Consume: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Server error"})
		return
	}

	for msg := range msgs {
		ws.WriteMessage(websocket.TextMessage, msg.Body)
	}
}

func getHome(c *gin.Context) {
	c.File("index.html")
}
