package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Maxxxxxx-x/gpx-downloader/internal/config"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Maxxxxxx-x/gpx-downloader/internal/downloader"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/logger"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/models"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/parser"
)

const (
	OUTPUT_PATH = "data-sources/gpx/"
    DOWNLOAD_FILES = false
)

func ensureDownloadPath() (string, error) {
	exec, err := os.Executable()
	if err != nil {
		return "", err
	}
	fullPath := path.Join(path.Dir(exec), "..", "..", OUTPUT_PATH)
	info, err := os.Stat(fullPath)
	if err == nil && info.IsDir() {
		return fullPath, nil
	}

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", err
	}

	return fullPath, nil
}

func main() {
	cfg, err := config.GetConfig("downloader")
	log := logger.New(cfg.Logging, cfg.Env)
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("Failed to get configurations")
	}

	log.Info().Msgf("Database enabled: %v", cfg.Database.Enabled)

    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Username,
        cfg.Database.Password,
        cfg.Database.DatabaseName,
    )

    dbconfig,err := pgxpool.ParseConfig(connStr)
    if err != nil {
        log.Fatal().Err(err).Msgf("Failed to connect to database")
    }

    dbconfig.MaxConns = 20
    dbconfig.MinConns = 10
    dbconfig.MaxConnIdleTime = 10 * time.Hour
    dbconfig.MaxConnLifetime = 10 * time.Hour
    dbconfig.MaxConnLifetimeJitter = 11 * time.Hour

    connPool, err := pgxpool.NewWithConfig(context.Background(), dbconfig)
    if err != nil {
		log.Fatal().Err(err).Msg("Unable to create pgx pool")
	}
	log.Info().Msg("Database pool created!")

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to db")
	}

	defer connPool.Close()

	installPath, err := ensureDownloadPath()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to ensure download path. Exiting")
		return
	}
	database := db.New(connPool, installPath, log)

	startTime := time.Now()

	var csvFiles []models.CSVFile
    if cfg.Env == "dev" {
        csvFiles, _ = parser.StartParser("data-sources/csv", log)
    } else {
        csvFiles, _ = parser.StartParser("/data-sources/csv", log)
    }

	if cfg.Database.Enabled {
		database.SaveCSVFilesToDatabase(csvFiles)
	}

	var recordsToDownload []*models.DataRecord
	for _, csvFile := range csvFiles {
		if csvFile.Data == nil {
			continue
		}
		recordsToDownload = append(recordsToDownload, csvFile.Data...)
        log.Info().Msgf("File: %s | Records: %d", csvFile.FileName, len(csvFile.Data))
	}
    log.Info().Msgf("Total records records expected: %d | Total files expected: %d", len(recordsToDownload), len(recordsToDownload) + len(csvFiles))

	if len(recordsToDownload) < 1 {
		log.Fatal().Msg("FAILED TO GET RECORDS TO DOWNLOAD. EXITIING")
		return
	}

    if DOWNLOAD_FILES {
	    downloader.StartDownload(recordsToDownload, installPath, log)
    }

	if cfg.Database.Enabled {
		database.SaveRecordsToDatabase(recordsToDownload)
	}

	log.Info().Msgf("All process completed! Total elapsed time: %v", time.Since(startTime))
}
