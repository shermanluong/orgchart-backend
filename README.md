# Backend(Go)

## Description
This backend service provides the organizational chart for the company by retrieving employee data from a remote API, 
storing it in a PostgreSQL database, and serving it via an API. The service allows for hierarchical representation of employees, sorted by their last name.

## Technologies Used
 - Go (Golang)
 - Gin Web Framework
 - GORM(ORM for PostgreSQL)
 - CORS(Cross-Origin Resource Sharing) handling

 ## Setup Instructions
 1. Clone the Repository.
 ```
 bash
 git clone https://github.com/leetechguru/orgchart-backend.git
 cd orgchart-backend
 ``` 
 2. Install Dependencies. Ensure Go is installed, and then install necessary dependencies:
 ```
 bash
 go mod tidy
 ```
 3. Setup PostgreSQL Database. Make sure you have PostgreSQL installed and a database created.
 
 Example for PostgreSQL setup
 ```
 bash
 createdb orgchart_db
 ```

 Update the database credentials in the backend code:
 ```
 go
 dsn := "host=localhost user=postgres password=123 dbname=orgchart_db port=5432 sslmode=disable TimeZone=UTC"
 ```
 4. Run the server. Start the backend service by running:
 ```
 bash
 go run main.go
 ```
 The backend API will be available at ```http://localhost:8080/org-chart```.

 ## API Endpoints
 - GET /org-chart
    - Description: Retrieves the organizational chart in JSON format.
    - Response: JSON data representing the organization, including full name, title, and reports(if any)

## CORS Handling
The backend is configured to handle CORS, allowing requests from frontend applications running on different origins.