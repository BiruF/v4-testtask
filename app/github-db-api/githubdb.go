package githubdb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v39/github"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

type Database struct {
	db *gorm.DB
}

func NewDatabase() (*Database, error) {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_PASSWORD"), os.Getenv("DATABASE_NAME"),
		os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_SSL"),
		os.Getenv("DATABASE_TIMEZONE"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	err = db.AutoMigrate(&Repository{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) FetchGitHubData(client *github.Client, username string) error {
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

		d.db.Create(&repository)
	}

	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) ReadAllRecords() ([]Repository, error) {
	var records []Repository
	if err := d.db.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}
