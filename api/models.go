package main

type User struct {
	ID        int64  `json:"id"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthdate string `json:"birthdate"`
	Gender    string `json:"gender"`
	Interests string `json:"interests"`
	City      string `json:"city"`
}

type PrintableUser struct {
	ID        int64  `json:"id"`
	Password  string `json:"-"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthdate string `json:"birthdate"`
	Gender    string `json:"gender"`
	Interests string `json:"interests"`
	City      string `json:"city"`
}

type Post struct {
	ID      int64  `json:"id"`
	Text    string `json:"text"`
	UserID  int64  `json:"user_id"`
	Updated string `json:"updated"`
}

type FeedUpdateInitiator struct {
	PostID   int64 `json:"post_id"`
	FriendID int64 `json:"friend_id"`
}

type FeedUpdateReceivers struct {
	PostID  int64   `json:"post_id"`
	UserIDs []int64 `json:"user_ids"`
}
