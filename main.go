// Package homie contains a simple homelab monitoring tool
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

var (
	DBHost        = os.Getenv("HOMIE_DB_HOST")
	DBPort        = os.Getenv("HOMIE_DB_PORT")
	DBUser        = os.Getenv("HOMIE_DB_USER")
	DBPassword    = os.Getenv("HOMIE_DB_PASSWORD")
	DBName        = os.Getenv("HOMIE_DB_NAME")
	CheckInterval, _ = strconv.Atoi(os.Getenv("HOMIE_INTERVAL"))
)

type Service struct {
	Name string
	URL  string
}

type ServiceStatus struct {
	Name         string
	URL          string
	Status       string
	ResponseTime time.Duration
	StatusCode   int
	CheckedAt    time.Time
	Error        string
}

func CheckURL(name string, url string) *ServiceStatus {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	status := &ServiceStatus{
		Name:         name,
		URL:          url,
		Status:       "ERR",
		ResponseTime: 0 * time.Second,
		StatusCode:   0,
		CheckedAt:    time.Now(),
		Error:        "",
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error preparing request: %v", err)
		status.Error = fmt.Sprintf("%v", err)
		return status
	}

	req.Header.Add("Accept", "application/json")

	startTime := time.Now()
	res, err := client.Do(req)
	durationTime := time.Since(startTime)
	status.ResponseTime = durationTime
	if err != nil {
		log.Printf("Error sending request: %v", err)
		status.Error = fmt.Sprintf("%v", err)
		return status
	}
	defer res.Body.Close()

	status.StatusCode = res.StatusCode

	if res.StatusCode != http.StatusOK {
		log.Printf("Ping failed with code: %d", res.StatusCode)
		status.Status = "DOWN"
		status.Error = fmt.Sprintf("HTTP %d", res.StatusCode)
		return status
	}

	status.Status = "UP"
	log.Printf("Check successful (%s) - %d", url, res.StatusCode)
	return status
}

func SaveCheckToDB(db *sql.DB, status *ServiceStatus) {
	var serviceID int
	err := db.QueryRow("SELECT id FROM services WHERE name = $1;", status.Name).Scan(&serviceID)
	if err != nil {
		log.Fatalf("Service not found: %v", status.Name)
	}
	_, err = db.Exec(
		"INSERT INTO health_checks (service_id, status, response_time, status_code, error_message) VALUES ($1, $2, $3, $4, $5)",
		serviceID, status.Status, status.ResponseTime, status.StatusCode, status.Error,
	)
	if err != nil {
		log.Fatalf("Error saving check to database: %v", err)
	}
}

func ReadServicesFromDB(db *sql.DB) []Service {
	services := []Service{}
	rows, err := db.Query("SELECT name, url FROM services;")
	if err != nil {
		log.Fatalf("Could not fetch services from the database: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var svc Service
		err = rows.Scan(&svc.Name, &svc.URL)
		if err != nil {
			log.Fatalf("Error parsing service row from database: %v", err)
		}
		services = append(services, svc)
	}
	return services
}

func main() {
	log.Print("Homie homelab monitoring tool")
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DBHost, DBPort, DBUser, DBPassword, DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	log.Println("Checking database connection")
	if err := db.Ping(); err != nil {
		log.Fatalf("Database not reachable: %v", err)
	}
	log.Println("Database connection healthy!")

	ticker := time.NewTicker(time.Duration(CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			services := ReadServicesFromDB(db)
			statuses := []*ServiceStatus{}
			for _, service := range services {
				statuses = append(statuses, CheckURL(service.Name, service.URL))
			}

			for _, check := range statuses {
				SaveCheckToDB(db, check)
				fmt.Printf("%s\t\t(%s)\t\tStatus:%s\t ResponseTime: %d ms\t StatusCode: %d\t CheckedAt: %v\t Error:%s\n", check.Name, check.URL, check.Status, check.ResponseTime.Milliseconds(), check.StatusCode, check.CheckedAt, check.Error)
			}
		}
	}
}
