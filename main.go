// Package homie contains a simple homelab monitoring tool
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

const (
	ImmichURL    = "https://photos.lazydev.sh/api/server/ping"
	BitwardenURL = "https://bitwarden.lazydev.sh/"
	InvalidURL   = "https://eva.cuate.pl/"
	DBHost       = "192.168.1.120"
	DBPort       = 5432
	DBUser       = "postgres"
	DBPassword   = "zaq12wsx"
	DBName       = "homie"
)

var (
	Services = []Service{
		{"Immich", ImmichURL},
		{"Bitwarden", BitwardenURL},
		{"Invalid", InvalidURL},
	}
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
	db.Exec(
		"INSERT INTO health_checks (service_id, status, response_time, status_code, error_message) VALUES ($1, $2, $3, $4, $5)", 
		serviceID, status.Status, status.ResponseTime, status.StatusCode, status.Error,
	)
}

func main() {
	log.Print("Homie homelab monitoring tool")
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		DBHost, DBPort, DBUser, DBPassword, DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	statuses := []*ServiceStatus{}
	for _, service := range Services {
		statuses = append(statuses, CheckURL(service.Name, service.URL))
	}

	for _, check := range statuses {
		SaveCheckToDB(db, check)
		fmt.Printf("%s\t\t(%s)\t\tStatus:%s\t ResponseTime: %d ms\t StatusCode: %d\t CheckedAt: %v\t Error:%s\n", check.Name, check.URL, check.Status, check.ResponseTime.Milliseconds(), check.StatusCode, check.CheckedAt, check.Error)
	}
}
