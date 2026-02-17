package main

import (
	"fmt"
	"os"

	"kartcis-backend/config"
	"kartcis-backend/jobs"
	"kartcis-backend/routes"
)

func main() {
	// Connect to Database
	config.ConnectDB()

	// Start Background Jobs
	jobs.StartOrderExpiryJob()
	jobs.StartPaymentCheckerJob()
	jobs.StartEventExpiryJob()

	// Setup Router
	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	apiPrefix := os.Getenv("API_PREFIX")
	if apiPrefix == "" {
		apiPrefix = "/api/v1"
	}
	fmt.Printf("Server is running on port %s\n", port)
	fmt.Printf("Access API at: http://localhost:%s%s\n", port, apiPrefix)
	r.Run(":" + port)
}
