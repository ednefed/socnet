package main

type Message struct {
	ID         int64  `json:"-"`
	FromUserID int64  `json:"from_user_id"`
	ToUserID   int64  `json:"to_user_id"`
	Message    string `json:"message"`
	CreatedAt  string `json:"created_at"`
	Read       bool   `json:"-"`
}

type Operation struct {
	FromUserID int64   `json:"from_user_id"`
	ToUserID   int64   `json:"to_user_id"`
	Operation  string  `json:"operation"`
	IDs        []int64 `json:"ids"`
}
