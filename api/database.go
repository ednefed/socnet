package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func connectToDB(host string, port string, name string, username string, password string, SSLMode string) *sql.DB {
	dsn := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v", username, password, host, port, name, SSLMode)
	database, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal(err)
	}

	if err := database.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected to %v:%v/%v as %v", host, port, name, username)
	maxOpenConnections, err := strconv.ParseInt(getEnvVar("POSTGRESQL_MAX_OPEN_CONNECTIONS", "95"), 10, 0)

	if err != nil {
		log.Fatal(err)
	}

	database.SetMaxOpenConns(int(maxOpenConnections))
	return database
}

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

	if err := db2.QueryRow(query, id).Scan(&user.ID, &user.Password, &user.FirstName, &user.LastName, &user.Birthdate, &user.Gender, &user.Interests, &user.City); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("dbGetUserByID.QueryRow: %v", err)
			return user, err
		}
	}

	return user, nil
}

func dbGetUsersByFistAndLastName(firstName, lastName string) ([]PrintableUser, error) {
	query := "SELECT id, first_name, last_name, birthdate, gender, interests, city FROM public.users WHERE first_name like $1 || '%' and last_name like $2 || '%'"
	var users []PrintableUser
	rows, err := db3.Query(query, firstName, lastName)

	if err != nil {
		log.Printf("dbGetUsersByFistAndLastName.Query: %v", err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var user PrintableUser

		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Birthdate, &user.Gender, &user.Interests, &user.City); err != nil {
			log.Printf("dbGetUsersByFistAndLastName.Scan: %v", err)
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Printf("dbGetUsersByFistAndLastName.rows: %v", err)
		return nil, err
	}

	if len(users) == 0 {
		return nil, sql.ErrNoRows
	}

	return users, nil
}
