package main

type Dialog struct {
	FromUserID int64 `json:"from_user_id"`
	ToUserID   int64 `json:"to_user_id"`
	Unread     int64 `json:"unread"`
}

type Operation struct {
	FromUserID int64   `json:"from_user_id"`
	ToUserID   int64   `json:"to_user_id"`
	Operation  string  `json:"operation"`
	IDs        []int64 `json:"ids"`
}
