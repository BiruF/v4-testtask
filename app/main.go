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
	pingTimeEndStr := os.Getenv("PINGTIME_END")
	outputFilePath := "./output.json"
	outputFilePath2 := "./output2.json"

	db, err := githubdb.NewDatabase()
	if err != nil {
		panic(fmt.Errorf("failed to create database instance: %v", err))
	}
	defer db.Close()

	client := github.NewClient(nil)
	gitHubUsername := os.Getenv("GITHUB_USERNAME")

	fmt.Printf("Ping time = %s minute\n", pingTimeStr)
	if pingTimeEndStr != "" {
		fmt.Printf("Ping time end = %s minute\n", pingTimeEndStr)
	} else {
		panic("PINGTIME_END environment variable is not set")
	}

	pingTime, err := strconv.Atoi(pingTimeStr)
	if err != nil {
		panic(fmt.Errorf("failed to convert PINGTIME to integer: %v", err))
	}

	pingTimeEnd, err := strconv.Atoi(pingTimeEndStr)
	if err != nil {
		panic(fmt.Errorf("failed to convert PINGTIME_END to integer: %v", err))
	}

	if pingTime >= pingTimeEnd {
		panic("PINGTIME_END should be greater than PINGTIME")
	}

	go func() {
		for {
			time.Sleep(time.Duration(pingTime) * time.Minute)

			if err := db.FetchGitHubData(client, gitHubUsername); err != nil {
				fmt.Println("Failed to fetch GitHub data:", err)
			} else {
				fmt.Println("Data has been updated in the database")
			}
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(pingTimeEnd-pingTime) * time.Minute)

			if err := githubdb.WriteJSONFile(db, outputFilePath); err != nil {
				fmt.Println("Failed to write JSON data:", err)
			} else {
				fmt.Println("JSON data has been updated and written to", outputFilePath)
				if err := githubdb.ReadJSONFile(outputFilePath); err != nil {
					fmt.Println("Failed to read JSON file:", err)
				}
			}
		}
	}()

	go func() {
		for {
			time.Sleep(1 * time.Minute) // Проверяем раз в минуту

			// Определяем годы из переменных окружения
			startYearStr := os.Getenv("START_YEAR")
			endYearStr := os.Getenv("END_YEAR")

			startYear, err := strconv.Atoi(startYearStr)
			if err != nil {
				fmt.Println("Failed to convert START_YEAR to integer:", err)
				continue
			}

			endYear, err := strconv.Atoi(endYearStr)
			if err != nil {
				fmt.Println("Failed to convert END_YEAR to integer:", err)
				continue
			}

			// Получаем записи из базы данных
			records, err := db.ReadAllRecords()
			if err != nil {
				fmt.Println("Failed to read records from database:", err)
				continue
			}

			// Фильтруем записи по году обновления и записываем в файл output2.json
			var filteredRecords []githubdb.Repository
			for _, record := range records {
				if record.UpdateTs != nil {
					updateYear := record.UpdateTs.Year()
					if updateYear >= startYear && updateYear <= endYear {
						filteredRecords = append(filteredRecords, record)
					}
				}
			}

			if err := githubdb.WriteFilteredJSONFile(filteredRecords, outputFilePath2); err != nil {
				fmt.Println("Failed to write filtered JSON data:", err)
			} else {
				fmt.Println("Filtered JSON data has been written to", outputFilePath2)
				if err := githubdb.ReadJSONFile(outputFilePath2); err != nil {
					fmt.Println("Failed to read JSON file:", err)
				}
			}
		}
	}()

	select {}
}
