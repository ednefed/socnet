package main

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func dbAddUser(user User) (int64, error) {
	query := "INSERT INTO public.users(password, first_name, last_name, birthdate, gender, interests, city) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		log.Printf("dbAddUser.GenerateFromPassword: %v", err)
		return 0, err
	}

	var id int64
	if err := db.QueryRow(query, string(hash), user.FirstName, user.LastName, user.Birthdate, user.Gender, user.Interests, user.City).Scan(&id); err != nil {
		log.Printf("dbAddUser.QueryRow: %v", err)
		return 0, err
	}

	return id, nil
}

func dbGetUserByID(id int64) (User, error) {
	query := "SELECT id, password, first_name, last_name, birthdate, gender, interests, city FROM public.users WHERE id = $1"
	var user User

	if err := db.QueryRow(query, id).Scan(&user.ID, &user.Password, &user.FirstName, &user.LastName, &user.Birthdate, &user.Gender, &user.Interests, &user.City); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("dbGetUserByID.QueryRow: %v", err)
			return user, err
		}
	}

	return user, nil
}
