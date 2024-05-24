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
