package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	githubdb "app/github-db-api"

	"github.com/google/go-github/v39/github"
)

func main() {
	pingTimeStr := os.Getenv("PINGTIME")

	db, err := githubdb.NewDatabase()
	if err != nil {
		panic(fmt.Errorf("failed to create database instance: %v", err))
	}
	defer db.Close()

	client := github.NewClient(nil)
	gitHubUsername := os.Getenv("GITHUB_USERNAME")

	fmt.Printf("Ping time = %s minute\n", pingTimeStr)

	pingTime, err := strconv.Atoi(pingTimeStr)
	if err != nil {
		panic(fmt.Errorf("failed to convert PINGTIME to integer: %v", err))
	}

	for range time.Tick(time.Duration(pingTime) * time.Minute) {
		if err := db.FetchGitHubData(client, gitHubUsername); err != nil {
			fmt.Println("Failed to fetch GitHub data:", err)
			continue
		}

		records, err := db.ReadAllRecords()
		if err != nil {
			fmt.Println("Failed to read records:", err)
			continue
		}

		fmt.Println("All records:")
		for _, record := range records {
			fmt.Printf("ID: %d, Username: %s, Name: %s, URL: %s\n", record.ID, record.Username, record.Name, record.URL)
		}
	}
}
