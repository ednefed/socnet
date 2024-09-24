package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/tarantool/go-tarantool/v2"
)

var tt *tarantool.Connection

func connectToTarantool(host string, port string, username string) *tarantool.Connection {
	ttctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dialer := tarantool.NetDialer{
		Address: fmt.Sprintf("%s:%s", host, port),
		User:    username,
	}

	opts := tarantool.Opts{}

	connection, err := tarantool.Connect(ttctx, dialer, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected to tarantool %s:%s as %s", host, port, username)
	return connection
}

func tarantoolCreateMessage(message Message) (Message, error) {
	query := fmt.Sprintf("box.space.dialogs:auto_increment{%s, %s, '%s', '%s'}", strconv.FormatInt(message.FromUserID, 10), strconv.FormatInt(message.ToUserID, 10), message.Message, message.CreatedAt)

	if _, err := tt.Do(tarantool.NewEvalRequest(query)).Get(); err != nil {
		log.Printf("tarantoolCreateMessage: %v", err)
		return message, err
	}

	return message, nil
}

func tarantoolSelectFromDialogs(query string) ([]Message, error) {
	var messages []Message
	raw, err := tt.Do(tarantool.NewEvalRequest(query)).Get()

	if err != nil {
		log.Printf("tarantoolGetMessages: %v", err)
		return nil, err
	}

	data := raw[0].([]interface{})

	for _, item := range data {
		message := Message{
			FromUserID: int64(item.([]interface{})[1].(int8)),
			ToUserID:   int64(item.([]interface{})[2].(int8)),
			Message:    item.([]interface{})[3].(string),
			CreatedAt:  item.([]interface{})[4].(string),
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func tarantoolGetDialogMessages(userID int64, friendID int64) ([]Message, error) {
	query := fmt.Sprintf("return box.space.dialogs.index.from_to:select({%s, %s}, {iterator = 'REQ'})", strconv.FormatInt(userID, 10), strconv.FormatInt(friendID, 10))
	var messages []Message
	messagesFromTo, err := tarantoolSelectFromDialogs(query)

	if err != nil {
		log.Printf("tarantoolGetDialogMessages: %v", err)
		return nil, err
	}

	var messagesToFrom []Message

	if userID != friendID {
		query = fmt.Sprintf("return box.space.dialogs.index.from_to:select({%s, %s}, {iterator = 'REQ'})", strconv.FormatInt(friendID, 10), strconv.FormatInt(userID, 10))
		messagesToFrom, err = tarantoolSelectFromDialogs(query)

		if err != nil {
			log.Printf("tarantoolGetDialogMessages: %v", err)
			return nil, err
		}
	}

	messages = append(messagesFromTo, messagesToFrom...)

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt > messages[j].CreatedAt
	})

	return messages, nil
}
