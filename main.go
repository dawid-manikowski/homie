// Package homie contains a simple homelab monitoring tool
package main

import (
	"log"
	"os"
	"strconv"
)

var (
	DBHost        = os.Getenv("HOMIE_DB_HOST")
	DBPort        = os.Getenv("HOMIE_DB_PORT")
	DBUser        = os.Getenv("HOMIE_DB_USER")
	DBPassword    = os.Getenv("HOMIE_DB_PASSWORD")
	DBName        = os.Getenv("HOMIE_DB_NAME")
	CheckInterval, _ = strconv.Atoi(os.Getenv("HOMIE_INTERVAL"))
)

func main() {
	log.Print("Homie homelab monitoring tool")
	db := SetUpDatabaseConnection(DBHost , DBPort , DBUser , DBPassword , DBName)
	go StartMonitor(db)
	StartWebServer(db)
}

