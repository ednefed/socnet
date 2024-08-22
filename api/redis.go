package main

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/go-redis/redis/v9"
)

var feedCache *redis.Client

func connectToRedis(host string, port string, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})
	return client
}

func cacheGetFeedForUser(userId int64, offset int64, limit int64) ([]Post, error) {
	postStrings := feedCache.LRange(ctx, strconv.FormatInt(userId, 10), offset, offset+limit-1).Val()
	var post Post
	var posts []Post

	for _, postString := range postStrings {
		if err := json.Unmarshal([]byte(postString), &post); err != nil {
			log.Printf("cacheGetFeedForUser.json.Unmarshal: %v", err)
			return posts, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func cacheUpdateFeedForUser(userID int64, post Post) error {
	payload, err := json.Marshal(post)

	if err != nil {
		log.Printf("cachePubFeedUpdate.json.Marshal: %v", err)
		return err
	}

	if err := feedCache.LPush(ctx, strconv.FormatInt(userID, 10), payload).Err(); err != nil {
		log.Printf("cacheUpdateFeedForUser.LPush: %v", err)
		return err
	}

	userFeedSize := feedCache.LLen(ctx, strconv.FormatInt(userID, 10)).Val()

	for userFeedSize > feedSize {
		err := feedCache.RPop(ctx, strconv.FormatInt(userID, 10)).Err()

		if err != nil {
			log.Printf("cacheUpdateFeedForUser.RPop: %v", err)
			return err
		}

		userFeedSize--
	}

	return nil
}

func cachePubPrepareFeedUpdate(initiator FeedUpdateInitiator) error {
	payload, err := json.Marshal(initiator)

	if err != nil {
		log.Printf("cachePubPrepareFeedUpdate.json.Marshal: %v", err)
		return err
	}

	if err := feedCache.Publish(ctx, "feedUpdatePrepare", payload).Err(); err != nil {
		log.Printf("cachePubPrepareFeedUpdate.Publish: %v", err)
		return err
	}

	log.Printf("cachePubPrepareFeedUpdate: sent initiator: %v", initiator)

	return nil
}

func cachePubFeedUpdate(receivers FeedUpdateReceivers) error {
	payload, err := json.Marshal(receivers)

	if err != nil {
		log.Printf("cachePubFeedUpdate.json.Marshal: %v", err)
		return err
	}

	if err := feedCache.Publish(ctx, "feedUpdate", payload).Err(); err != nil {
		log.Printf("cachePubFeedUpdate.Publish: %v", err)
		return err
	}

	log.Printf("cachePubFeedUpdate: sent receivers: %v", receivers)

	return nil
}

func cacheSubPrepareFeedUpdate() {
	subscription := feedCache.PSubscribe(ctx, "feedUpdatePrepare").Channel()
	var initiator FeedUpdateInitiator

	for {
		msg := <-subscription

		if err := json.Unmarshal([]byte(msg.Payload), &initiator); err != nil {
			log.Printf("cacheSubPrepareFeedUpdate.json.Unmarshal: %v", err)
		}

		log.Printf("cacheSubPrepareFeedUpdate: rcvd initiator: %v", initiator)
		friendsCount, err := dbGetFriendsCountByUserID(initiator.FriendID)

		if err != nil {
			log.Printf("createPost.dbGetFriendsCount: %v", err)
		}

		var i int64
		var friends []int64

		for i = 0; i < friendsCount; i += feedUpdateBatchSize {
			friends, err = dbGetFriendsByUserID(initiator.FriendID, i, feedUpdateBatchSize)

			if err != nil {
				log.Printf("createPost.dbGetFriends: %v", err)
			}

			receivers := FeedUpdateReceivers{
				PostID:  initiator.PostID,
				UserIDs: friends,
			}

			err = cachePubFeedUpdate(receivers)

			if err != nil {
				log.Printf("createPost.cachePubFeedUpdate: %v", err)
			}
		}
	}
}

func cacheSubFeedUpdate() {
	subscription := feedCache.PSubscribe(ctx, "feedUpdate").Channel()
	var receivers FeedUpdateReceivers

	for {
		msg := <-subscription

		if err := json.Unmarshal([]byte(msg.Payload), &receivers); err != nil {
			log.Fatalf("cacheSubFeedUpdate.json.Unmarshal: %v", err)
		}

		log.Printf("cacheSubFeedUpdate: rcvd receivers: %v", receivers)
		post, err := dbGetPostByID(receivers.PostID)

		if err != nil {
			log.Printf("cacheSubFeedUpdate.dbGetPostByID: %v", err)
			continue
		}

		for _, userID := range receivers.UserIDs {
			err := cacheUpdateFeedForUser(userID, post)

			if err != nil {
				log.Printf("cacheSubFeedUpdate.cacheUpdateFeedForUser: %v", err)
				continue
			}
		}
	}
}
