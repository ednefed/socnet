package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/go-redis/redis/v9"
)

var cache *redis.Client

func connectToRedis(host string, port string, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})

	log.Printf("Connected to redis %v:%v", host, port)
	return client
}

func cacheGetUnreadMessagesCount(friendID, userID int64) (int64, error) {
	key := fmt.Sprintf("%s_%s", strconv.FormatInt(friendID, 10), strconv.FormatInt(userID, 10))
	countString, err := cache.Get(context.Background(), key).Result()

	if err != nil {
		if err == redis.Nil {
			return 0, nil
		} else {
			log.Printf("getUnreadMessagesCount.Get: %v", err)
			return 0, err
		}
	}

	count, err := strconv.ParseInt(countString, 10, 64)

	if err != nil {
		log.Printf("getUnreadMessagesCount.ParseInt: %v", err)
		return 0, err
	}

	return count, nil
}

func cacheIncrementUnreadMessagesCount(friendID, userID, value int64) error {
	key := fmt.Sprintf("%s_%s", strconv.FormatInt(friendID, 10), strconv.FormatInt(userID, 10))
	_, err := cache.IncrBy(context.Background(), key, value).Result()

	if err != nil {
		log.Printf("cacheIncrementUnreadMessagesCount.IncrBy: %v", err)
	} else {
		log.Printf("cacheIncrementUnreadMessagesCount: incremented counter for %s", key)
	}

	return err
}

func cacheDecrementUnreadMessagesCount(friendID, userID, value int64) error {
	key := fmt.Sprintf("%s_%s", strconv.FormatInt(friendID, 10), strconv.FormatInt(userID, 10))
	result, err := cache.DecrBy(context.Background(), key, value).Result()

	if err != nil {
		log.Printf("cacheDecrementUnreadMessagesCount.DecrBy: %v", err)
	} else {
		log.Printf("cacheDecrementUnreadMessagesCount: decremented counter for %s", key)
	}

	if result < 0 {
		_, err = cache.Set(context.Background(), key, 0, 0).Result()

		if err != nil {
			log.Printf("cacheDecrementUnreadMessagesCount.Set: %v", err)
		}
	}

	return err
}
