package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var citus *sql.DB

func citusCreateMessage(message Message) (Message, error) {
	query := "INSERT INTO public.dialogs(from_user, to_user, message) VALUES ($1, $2, $3)"
	var createdMessage Message

	if _, err := citus.Exec(query, message.FromUserID, message.ToUserID, message.Message); err != nil {
		log.Printf("citusCreateMessage.Exec: %v", err)
		return createdMessage, err
	}

	return createdMessage, nil
}

func citusGetDialogMessages(userID int64, friendID int64) ([]Message, error) {
	query := "SELECT from_user, to_user, message, created_at FROM public.dialogs WHERE (from_user = $1 AND to_user = $2) OR (from_user = $2 AND to_user = $1) ORDER BY created_at DESC"
	var messages []Message
	rows, err := citus.Query(query, userID, friendID)

	if err != nil {
		log.Printf("citusGetDialogMessages.Query: %v", err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var message Message

		if err := rows.Scan(&message.FromUserID, &message.ToUserID, &message.Message, &message.CreatedAt); err != nil {
			log.Printf("citusGetDialogMessages.Scan: %v", err)
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		log.Printf("citusGetDialogMessages.rows: %v", err)
		return nil, err
	}

	if len(messages) == 0 {
		return nil, sql.ErrNoRows
	}

	return messages, nil
}
