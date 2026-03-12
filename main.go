package main

import (
	"fmt"
	"os"
	_ "time/tzdata" // Embed IANA timezone database ke binary (tidak perlu tzdata di OS)

	"kartcis-backend/config"
	"kartcis-backend/jobs"
	"kartcis-backend/routes"
)

func main() {
	// Connect to Database
	config.ConnectDB()

	// Start Background Jobs
	jobs.StartOrderExpiryJob()
	jobs.StartEventExpiryJob()

	// Ensure uploads directory exists and has public read access for Nginx
	if err := os.MkdirAll("uploads", 0755); err != nil {
		fmt.Printf("Warning: failed to mkdki uploads: %v\n", err)
	}
	if err := os.Chmod("uploads", 0755); err != nil {
		fmt.Printf("Warning: failed to chmod uploads: %v\n", err)
	}

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
