package main

import (
	"time"
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

