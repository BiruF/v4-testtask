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
	pingtimestr := os.Getenv("PINGTIME")
	dbuser := os.Getenv("DATABASE_USER")
	dbpassword := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	dbhost := os.Getenv("DATABASE_HOST")
	dbport := os.Getenv("DATABASE_PORT")
	dbssl := os.Getenv("DATABASE_SSL")
	dbtimezone := os.Getenv("DATABASE_TIMEZONE")

	fmt.Printf("ping time = %s minut\n", pingtimestr)
	fmt.Printf("database host= %s\n", dbhost)
	fmt.Printf("database port= %s\n", dbport)

	pingtime, err := strconv.Atoi(pingtimestr)
	if err != nil {
		panic(fmt.Errorf("failed to load config: %v", err)) //1245
	}

	db, err := githubdb.ConnectToDatabase(dbuser, dbpassword, dbname, dbhost, dbport, dbssl, dbtimezone)
	if err != nil {
		panic(fmt.Errorf("failed to connect to the database: %v", err))
	}
	defer githubdb.CloseDatabase(db)

	config, err := githubdb.LoadConfig("config.json")
	if err != nil {
		panic(fmt.Errorf("failed to load config: %v", err))
	}

	client := github.NewClient(nil)

	err = githubdb.FetchGitHubData(client, db, config.GitHub.Username)
	if err != nil {
		fmt.Println("Failed to fetch GitHub data:", err)
	}

	records, err := githubdb.ReadAllRecords(db)
	if err != nil {
		fmt.Println("Failed to read records:", err)
	} else {
		fmt.Println("All records:")
		for _, record := range records {
			fmt.Printf("ID: %d, Username: %s, Name: %s, URL: %s\n", record.ID, record.Username, record.Name, record.URL)
		}
	}

	ticker := time.NewTicker(time.Duration(pingtime) * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := githubdb.FetchGitHubData(client, db, config.GitHub.Username)
			if err != nil {
				fmt.Println("Failed to fetch GitHub data:", err)
			}

			records, err := githubdb.ReadAllRecords(db)
			if err != nil {
				fmt.Println("Failed to read records:", err)
			} else {
				fmt.Println("All records:")
				for _, record := range records {
					fmt.Printf("ID: %d, Username: %s, Name: %s, URL: %s\n", record.ID, record.Username, record.Name, record.URL)
				}
			}
		}
	}
}
