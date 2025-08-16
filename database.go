package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func SetUpDatabaseConnection(DBHost string, DBPort string, DBUser string, DBPassword string, DBName string) *sql.DB {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DBHost, DBPort, DBUser, DBPassword, DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	log.Println("Checking database connection")
	if err := db.Ping(); err != nil {
		log.Fatalf("Database not reachable: %v", err)
	}
	log.Println("Database connection healthy!")

	return db
}
