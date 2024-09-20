package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var rabbitmq *amqp.Connection

func connectToRabbitMQ(host string, port string, username string, password string) *amqp.Connection {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port))

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected to rabbitmq %s:%s as %s", host, port, username)
	return conn
}

func rmqPubPost(post Post) error {
	channel, err := rabbitmq.Channel()

	if err != nil {
		log.Printf("rmqPubPost.Channel: %v", err)
		return err
	}

	defer channel.Close()
	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload, err := json.Marshal(post)

	if err != nil {
		log.Printf("rmqPubPost.json.Marshal: %v", err)
		return err
	}

	err = channel.PublishWithContext(context,
		"amq.direct",
		strconv.FormatInt(post.UserID, 10),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)

	if err != nil {
		log.Printf("rmqPubPost.Publish: %v", err)
		return err
	}

	return nil
}
