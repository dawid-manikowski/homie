package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

func StartMonitor(db *sql.DB) {
	ticker := time.NewTicker(time.Duration(CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
			case <- ticker.C:
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

func GetCurrentServicesStatuses(db *sql.DB) []ServiceStatus {
	var statuses = []ServiceStatus{}
	query := `
		SELECT s.name, hc.status, hc.response_time, hc.error_message, hc.checked_at
		FROM services s
		JOIN health_checks hc ON s.id = hc.service_id
		WHERE hc.checked_at = (
				SELECT MAX(checked_at) 
				FROM health_checks hc2 
				WHERE hc2.service_id = hc.service_id
		);
	`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error fetching current services statuses: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var svcStatus ServiceStatus
		err = rows.Scan(&svcStatus.Name, &svcStatus.Status, &svcStatus.ResponseTime, &svcStatus.Error, &svcStatus.CheckedAt)	
		if err != nil {
			log.Printf("Could not unpack row: %v", err)
			continue
		} 
		statuses = append(statuses, svcStatus)
	}
	return statuses
}
