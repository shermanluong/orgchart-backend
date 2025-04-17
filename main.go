package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Employee struct represents an employee in the database
type Employee struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"` // Full name from JSON
	Title     string `json:"title"`
	ManagerID *int   `json:"manager_id"`
}

// OrgNode represents an employee in the organizational hierarchy
type OrgNode struct {
	FullName string    `json:"full_name"`
	Title    string    `json:"title"`
	Reports  []OrgNode `json:"reports,omitempty"`
}

var db *gorm.DB

func main() {
	// Database configuration (Update credentials as needed)
	dsn := "host=localhost user=postgres password=123 dbname=orgchart_db port=5432 sslmode=disable TimeZone=UTC"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL database:", err)
	}

	db.AutoMigrate(&Employee{})

	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/org-chart", getOrgChart)
	r.Run(":8080")
}

// getOrgChart retrieves and returns the organizational chart as JSON
func getOrgChart(c *gin.Context) {
	var count int64
	db.Model(&Employee{}).Count(&count)

	if count == 0 {
		if err := fetchAndStoreEmployees(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employee data"})
			return
		}
	}

	var employees []Employee
	db.Find(&employees)

	orgChart := buildHierarchy(employees)
	c.JSON(http.StatusOK, orgChart)
}

// fetchAndStoreEmployees retrieves employee data from an external API and stores it in the database
func fetchAndStoreEmployees() error {
	url := "https://gist.githubusercontent.com/chancock09/6d2a5a4436dcd488b8287f3e3e4fc73d/raw/fa47d64c6d5fc860fabd3033a1a4e3c59336324e/employees.json"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()

	var employees []Employee
	if err := json.NewDecoder(resp.Body).Decode(&employees); err != nil {
		fmt.Println(err)
		return err
	}

	for _, emp := range employees {
		db.Create(&emp)
	}
	return nil
}

// buildHierarchy constructs an organizational hierarchy from a list of employees
func buildHierarchy(employees []Employee) []OrgNode {
	empMap := make(map[int]*OrgNode)
	managerToReports := make(map[int][]int) // store employee IDs of reports

	// Split the Full Name into First Name and Last Name
	for _, emp := range employees {
		// Split name into first and last
		parts := strings.SplitN(emp.Name, " ", 2)
		firstName := parts[0]
		lastName := ""
		if len(parts) > 1 {
			lastName = parts[1]
		}

		empMap[emp.ID] = &OrgNode{
			FullName: fmt.Sprintf("%s %s", firstName, lastName),
			Title:    emp.Title,
		}

		if emp.ManagerID != nil {
			managerToReports[*emp.ManagerID] = append(managerToReports[*emp.ManagerID], emp.ID)
		}
	}

	var buildReports func(managerID int) []OrgNode
	buildReports = func(managerID int) []OrgNode {
		childIDs := managerToReports[managerID]

		// Sort by Last Name instead of Full Name
		sort.Slice(childIDs, func(i, j int) bool {
			// Split the Full Name into first and last names
			child1 := empMap[childIDs[i]].FullName
			child2 := empMap[childIDs[j]].FullName

			parts1 := strings.SplitN(child1, " ", 2)
			parts2 := strings.SplitN(child2, " ", 2)

			// Default to empty last name if it's not present
			lastName1 := ""
			if len(parts1) > 1 {
				lastName1 = parts1[1]
			}

			lastName2 := ""
			if len(parts2) > 1 {
				lastName2 = parts2[1]
			}

			// Compare by last name
			return lastName1 < lastName2
		})

		var reports []OrgNode
		for _, childID := range childIDs {
			childNode := empMap[childID]
			childNode.Reports = buildReports(childID)
			reports = append(reports, *childNode)
		}
		return reports
	}

	var topLevel []OrgNode
	for _, emp := range employees {
		if emp.ManagerID == nil {
			rootNode := empMap[emp.ID]
			rootNode.Reports = buildReports(emp.ID)
			topLevel = append(topLevel, *rootNode)
		}
	}

	// Sort top-level managers by last name as well
	sort.Slice(topLevel, func(i, j int) bool {
		// Split the Full Name into first and last names
		parts1 := strings.SplitN(topLevel[i].FullName, " ", 2)
		parts2 := strings.SplitN(topLevel[j].FullName, " ", 2)

		lastName1 := ""
		if len(parts1) > 1 {
			lastName1 = parts1[1]
		}

		lastName2 := ""
		if len(parts2) > 1 {
			lastName2 = parts2[1]
		}

		// Compare by last name
		return lastName1 < lastName2
	})

	return topLevel
}
