package main

import "os"

func main() {
	a := App{}
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8010"
	}

	a.Run(":" + port)
}
