package githubdb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v39/github"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		SSLMode  string `json:"sslmode"`
		TimeZone string `json:"timezone"`
	} `json:"database"`
	GitHub struct {
		Username string `json:"username"`
	} `json:"github"`
}

type Repository struct {
	ID          uint       `gorm:"primaryKey"`
	Ts          time.Time  `gorm:"column:ts"`
	Username    string     `gorm:"column:username"`
	Name        string     `gorm:"column:name"`
	UpdateTs    *time.Time `gorm:"column:update_ts"`
	Size        int        `gorm:"column:size"`
	Description string     `gorm:"column:description"`
	URL         string     `gorm:"column:url"`
}

func ConnectToDatabase(dbuser, dbpassword, dbname, dbhost, dbport, dbssl, dbtimezone string) (*gorm.DB, error) {
	/*config, err := LoadConfig("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	*/

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s TimeZone=%s",
		dbuser, dbpassword, dbname, dbhost, dbport, dbssl, dbtimezone)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	err = db.AutoMigrate(&Repository{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return db, nil
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	return &config, nil
}

func FetchGitHubData(client *github.Client, db *gorm.DB, username string) error {
	repos, _, err := client.Repositories.List(context.Background(), username, nil)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		var updateTs *time.Time
		if !repo.GetUpdatedAt().IsZero() {
			updatedAt := time.Time(repo.GetUpdatedAt().Time)
			updateTs = &updatedAt
		}

		parts := strings.Split(repo.GetFullName(), "/")
		repository := Repository{
			Username:    username,
			Name:        parts[1],
			URL:         repo.GetHTMLURL(),
			Description: repo.GetDescription(),
			UpdateTs:    updateTs,
			Ts:          time.Now(),
		}

		db.Create(&repository)
	}

	return nil
}

func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func ReadAllRecords(db *gorm.DB) ([]Repository, error) {
	var records []Repository
	if err := db.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}
