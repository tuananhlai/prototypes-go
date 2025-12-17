package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	globalCtx := context.Background()
	ctx, cancel := context.WithTimeout(globalCtx, 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}

	log.Println("== Create Table ==")
	schema := `CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		age INTEGER,
		bio TEXT
	);`
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("== Insert Rows ==")
	res, err := db.Exec("INSERT INTO users (username, age, bio) VALUES (?, ?, ?), (?, ?, ?)", "johndoe", 30, "Software Engineer", "janedoe", 25, "Product Owner")
	if err != nil {
		log.Fatal(err)
	}
	lastID, _ := res.LastInsertId()
	log.Printf("Inserted record with ID: %d\n", lastID)

	log.Println("== Select All Rows ==")
	rows, err := db.Query("SELECT id, username, age FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username string
		var age int
		if err := rows.Scan(&id, &username, &age); err != nil {
			log.Fatal(err)
		}
		log.Printf("User: %d, %s, %d\n", id, username, age)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	log.Println("== Select Row with ID 1 ==")
	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", 1).Scan(&username)
	if err == sql.ErrNoRows {
		log.Println("No user found with ID 1")
	} else if err != nil {
		log.Fatal(err)
	}
	log.Println("Username:", username)

	log.Println("== Insert Row with NULL field ==")
	_, _ = db.Exec("INSERT INTO users (username, age, bio) VALUES (?, ?, ?)", "ghost", nil, nil)

	var bio sql.NullString
	var age sql.NullInt64
	err = db.QueryRow("SELECT age, bio FROM users WHERE username = ?", "ghost").Scan(&age, &bio)
	if err != nil {
		log.Fatal(err)
	}
	if bio.Valid {
		log.Println("Bio:", bio.String)
	} else {
		fmt.Println("Bio is NULL")
	}

	log.Println("== Prepare Statement ==")
	stmt, err := db.Prepare("SELECT username FROM users WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var name string
	_ = stmt.QueryRow(1).Scan(&name)
	log.Println("Username:", name)

	log.Println("== Run Transaction ==")
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	{
		_, err := tx.Exec("UPDATE users SET age = age + 1 WHERE id = ?", 1)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatal(err)
		}
		log.Println("Transaction completed.")
	}

	log.Println("== Get Database Stats ==")
	stats := db.Stats()
	log.Printf("Open connections: %d\n", stats.OpenConnections)

	log.Println("== List Column Types ==")
	rows, _ = db.Query("SELECT * FROM users")
	cols, _ := rows.ColumnTypes()
	for _, col := range cols {
		log.Printf("Column name: %s, Type: %s\n", col.Name(), col.DatabaseTypeName())
	}
	rows.Close()
}
