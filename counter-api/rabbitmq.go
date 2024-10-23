package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var queue *amqp.Connection

func connectToRabbitMQ(host string, port string, username string, password string) *amqp.Connection {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port))

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected to rabbitmq %s:%s as %s", host, port, username)
	return conn
}

func queueSubscribeForOperations() {
	channel, err := queue.Channel()

	if err != nil {
		log.Fatalf("queueSubscribeForOperations.Channel: %v", err)
	}

	defer channel.Close()
	queue, err := channel.QueueDeclare(
		"DialogCounterIn",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("queueSubscribeForOperations.QueueDeclare: %v", err)
	}

	err = channel.QueueBind(
		queue.Name,
		"DialogCounterIn",
		"amq.direct",
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("queueSubscribeForOperations.QueueBind: %v", err)
	}

	messages, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("queueSubscribeForOperations.Consume: %v", err)
	}

	var operation Operation

	for message := range messages {
		if err := json.Unmarshal(message.Body, &operation); err != nil {
			log.Fatalf("queueSubscribeForOperations.json.Unmarshal: %v", err)
		}

		log.Printf("queueSubscribeForOperations: Received operation from %s: %v", queue.Name, operation)

		switch operation.Operation {
		case "increment":
			if err := cacheIncrementUnreadMessagesCount(operation.FromUserID, operation.ToUserID, int64(len(operation.IDs))); err != nil {
				log.Fatalf("queueSubscribeForOperations.cacheIncrementUnreadMessagesCount: %v", err)
			} else {
				message.Ack(false)
			}
		case "decrement":
			if err := cacheDecrementUnreadMessagesCount(operation.FromUserID, operation.ToUserID, int64(len(operation.IDs))); err != nil {
				operation.Operation = "decrement_failure"

				if err := queuePublishOperation(operation); err != nil {
					log.Printf("createDialogMessage.queuePublishOperation: %v", err)
				}

				message.Ack(false)
				log.Fatalf("queueSubscribeForOperations.cacheDecrementUnreadMessagesCount: %v", err)
			} else {
				message.Ack(false)
			}
		default:
			log.Println("queueSubscribeForOperations: unknown operation")
			message.Ack(false)
		}
	}
}

func queuePublishOperation(operation Operation) error {
	channel, err := queue.Channel()

	if err != nil {
		log.Printf("queuePublishOperation.Channel: %v", err)
		return err
	}

	defer channel.Close()
	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	payload, err := json.Marshal(operation)

	if err != nil {
		log.Printf("queuePublishOperation.json.Marshal: %v", err)
		return err
	}

	queueName := "DialogCounterOut"

	err = channel.PublishWithContext(context,
		"amq.direct",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)

	if err != nil {
		log.Printf("queuePublishOperation.Publish: %v", err)
		return err
	}

	log.Printf("queuePublishOperation: Published operation to %s: %v", queueName, operation)

	return nil
}
