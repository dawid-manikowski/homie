package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

var (
	tmpl *template.Template
)

const htmlTemplate = `
<html>
	<meta http-equiv="refresh" content="30">
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}

		body {
			background-color: #1a1a1a;
			color: #e0e0e0;
		}

		.add-service {
			margin: auto;
			width: 20%;
		}

		table {
			width: 100%;
			border-collapse: collapse;
			background-color: #2d2d2d;
			border-radius: 8px;
			overflow: hidden;
			box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
			margin-bottom: 100px;
		}

		th {
			background-color: #3d3d3d;
			color: #ffffff;
			padding: 15px;
			text-align: left;
			font-weight: 600;
			border-bottom: 2px solid #4a4a4a;
		}

		td {
			padding: 12px 15px;
			border-bottom: 1px solid #404040;
		}

		tr:hover {
			background-color: #363636;
		}

		.status-badge {
			padding: 4px 12px;
			border-radius: 20px;
			font-weight: bold;
			font-size: 0.85rem;
			text-transform: uppercase;
		}

		.UP {
			background-color: #10b981;
			color: #ffffff;
		}

		.DOWN {
			background-color: #ef4444;
			color: #ffffff;
		}

		.ERR {
			background-color: #f59e0b;
			color: #ffffff;
		}
	</style>
	<div>
		<table>
		<tr>
			<th>Service</th>
			<th>Status</th>
			<th>Last Checked</th>
			<th>Response Time</th>
		</tr>
		{{ range .Services}}
		<tr>
			<td>{{.Name}}</td>
			<td><span class="status-badge {{.Status}}">{{.Status}}</span></td>
			<td>{{.CheckedAt}}</td>
			<td>{{.ResponseTime}}</td>
		</tr>
		{{end}}
		</table>
	</div>
	<div class="add-service">
		<form action="/services" method="POST">
			<label for="service-name">Service Name</label><br />
			<input name="service-name" id="service-name" value="Sample" /><br /><br />

			<label for="service-address">Service Address</label><br />
			<input name="service-address" id="service-address" value="https://sample.com" /><br /><br />

			<button>Add</button>
		</form>
	</div>
</html>
`

func StartWebServer(db *sql.DB) {
	tmpl = template.Must(template.New("dashboard").Parse(htmlTemplate))
	http.HandleFunc("/", MainPage(db))
	http.HandleFunc("/services", Services(db))
	log.Println("Web server starting")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("Webserver failure: %v", err)
	}
}

func MainPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		services := GetCurrentServicesStatuses(db)
		data := DashboardData{services}
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
		}
	}
}

func Services(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			log.Print("Listing services")
		case http.MethodPost:
			log.Print("Adding service")
			svc := Service{r.FormValue("service-name"), r.FormValue("service-address")}
			log.Printf("%v", svc)
			err := SaveServiceToDB(db, &svc)
			if err != nil {
				log.Printf("Failed to add service: %v", err)
			}
		}
	}
}
