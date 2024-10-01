package main

type Message struct {
	FromUserID int64  `json:"from_user_id"`
	ToUserID   int64  `json:"to_user_id"`
	Message    string `json:"message"`
	CreatedAt  string `json:"created_at"`
}
