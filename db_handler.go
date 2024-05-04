package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func createDatabase() {
	_, err := os.Stat("db.sqlite")
	if err == nil {
		fmt.Println("Databse already exists")
		return
	}

	db, err := sql.Open("sqlite3", "db.sqlite")
	if err != nil {
		fmt.Printf("error creating db, %v\n", err)
		return
	}
	defer db.Close()

	q := "create table User (id integer not null primary key autoincrement, username text, password text)"
	_, err = db.Exec(q)
	if err != nil {
		fmt.Printf("error executing create... %v\n", err)
	}

	createUser(db, "ziga", "qwe")
	createUser(db, "qwe", "qwe")
	createUser(db, "venom", "venom")
	createUser(db, "test", "test")
	createUser(db, "u", "")
	createUser(db, "user", "user")
	createUser(db, "admin", "admin")

}

func createUser(db *sql.DB, user, password string) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("error creating tx")
	}
	stmt, err := tx.Prepare("insert into User(username, password) values (?,?)")
	if err != nil {
		fmt.Printf("error creating statement, %v\n", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user, hashPassword(password))
	if err != nil {
		fmt.Println("error execuring statement")
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("error commiting insert")
	}
}

func getUser(username, password string) bool {
	db, err := sql.Open("sqlite3", "db.sqlite")

	stmt, err := db.Prepare("select username, password from User where username = ?")
	if err != nil {
		fmt.Printf("error creating statement, %v\n", err)
		return false
	}
	defer stmt.Close()
	var u, passwordHashDb string
	err = stmt.QueryRow(username).Scan(&u, &passwordHashDb)
	if err != nil {
		fmt.Printf("error query row, %v\n", err)
		return false
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHashDb), []byte(password)) != nil {
		fmt.Printf("Login failed for user '%s' (worng password)\n", username)
		return false
	}

	fmt.Printf("Login succesful for user '%s'\n", username)
	return true
}

func hashPassword(password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return []byte{}
	}
	return hash
}
