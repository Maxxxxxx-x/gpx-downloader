package db

import (
	"context"
	"math"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/Maxxxxxx-x/gpx-downloader/internal/logger"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/models"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/sql/sqlc"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/ulid"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Database interface {
	SaveCSVFilesToDatabase(csvFile []models.CSVFile)
	SaveRecordsToDatabase(records []*models.DataRecord)
}

type BaseDatabase struct {
	log         zerolog.Logger
	installPath string
	queries     *sqlc.Queries
}

func createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Minute)
}

func (db *BaseDatabase) saveFilesToDatabase(filesToInsert []sqlc.BulkInsertFilesParams) {
	db.log.Info().Msgf("Inserting %d files into the database...", len(filesToInsert))
	ctx, cancel := createContext()
	defer cancel()
    defer func() {
        if r := recover(); r != nil {
            for idx := range filesToInsert {
                newId, err := ulid.GenerateULID()
                if err != nil {
                    db.log.Error().Err(err).Msg("Failed to regenerate file id")
                    return
                }
                filesToInsert[idx].ID = newId
            }
             db.saveFilesToDatabase(filesToInsert)
        }
    }()
	rowsAffected, err := db.queries.BulkInsertFiles(ctx, filesToInsert)
	if err != nil {
        db.log.Panic().Err(err).Send()
	} else {
		db.log.Info().Msgf("Saved %d files | Rows affected: %d", len(filesToInsert), rowsAffected)
	}
}

func (db *BaseDatabase) saveUserToDatabase(userId string) {
	ctx, cancel := createContext()
	defer cancel()
	if err := db.queries.InsertUser(ctx, userId); err != nil {
		db.log.Error().Err(err).Send()
	} else {
		db.log.Debug().Msgf("Saved user with ID %s", userId)
	}
}

func (db *BaseDatabase) saveRecordsToDatabase(records []sqlc.BulkInsertRecordParams) {
	db.log.Info().Msgf("saving %d records to database", len(records))
	ctx, cancel := createContext()
	defer cancel()
	rowsAffected, err := db.queries.BulkInsertRecord(ctx, records)
	if err != nil {
		db.log.Error().Err(err).Send()
	} else {
		db.log.Info().Msgf("Saved %d records | Rows affected: %d", len(records), rowsAffected)
	}
}

func (db *BaseDatabase) SaveCSVFilesToDatabase(csvFiles []models.CSVFile) {
	db.log.Info().Msgf("Preparing %d CSV file records", len(csvFiles))
	var insertParams []sqlc.BulkInsertFilesParams
	for _, file := range csvFiles {
		id, err := ulid.GenerateULID()
		if err != nil {
			db.log.Error().Err(err).Send()
			continue
		}
		param := sqlc.BulkInsertFilesParams{
			ID:        id,
			Filename:  file.FileName,
			Sha512sum: file.SHA512Sum,
		}
		insertParams = append(insertParams, param)
		db.log.Info().Msgf("Files prepared: %d / %d", len(insertParams), len(csvFiles))
	}

	db.saveFilesToDatabase(insertParams)
}

func (db *BaseDatabase) SaveRecordsToDatabase(records []*models.DataRecord) {
	batchSize := 1500
	skipped := 0
	recordCount := len(records)
	batchCount := int(math.Ceil(float64(recordCount / batchSize)))

	for i := 0; i <= batchCount; i++ {
		lower := skipped
		upper := skipped + batchSize
		if upper > recordCount {
			upper = recordCount
		}
		batchRecords := records[lower:upper]
		skipped += batchSize

        var preparedRecords []sqlc.BulkInsertRecordParams
        recordChan := make(chan sqlc.BulkInsertRecordParams, batchSize)

        var preparedFiles []sqlc.BulkInsertFilesParams
        filesChan := make(chan sqlc.BulkInsertFilesParams, batchSize)

        usersChan := make(chan string, batchSize)

        startTime := time.Now()
        go func() {
			for record := range recordChan {
				preparedRecords = append(preparedRecords, record)
			}
		}()

		go func() {
			for file := range filesChan {
				preparedFiles = append(preparedFiles, file)
			}
		}()

		go func() {
			for user := range usersChan {
				db.saveUserToDatabase(user)
			}
		}()

		var wg sync.WaitGroup
		for _, record := range batchRecords {
			go func(record *models.DataRecord) {
				wg.Add(1)
				defer wg.Done()
				fileId, err := ulid.GenerateULID()
				if err != nil {
					db.log.Error().Err(err).Msgf("Failed to generate file id for %s", record.FileName)
					return
				}

				recordId, err := ulid.GenerateULID()
				if err != nil {
					db.log.Error().Err(err).Msg("Failed to generate record id for record")
					return
				}

				filePath := path.Join(db.installPath, record.FileName)
				file, err := os.Open(filePath)
				if err != nil {
					db.log.Error().Err(err).Msgf("Failed to open file %s", filePath)
					return
				}

				fileHash, err := utils.GenerateFileHash(file)
				if err != nil {
					db.log.Error().Err(err).Msgf("Failed to generate file hash for file %s", file.Name())
					return
				}

				fileToInsert := sqlc.BulkInsertFilesParams{
					ID:        fileId,
					Filename:  record.FileName,
					Sha512sum: fileHash,
				}

				filesChan <- fileToInsert

				fileData, err := os.ReadFile(filePath)
				if err != nil {
					db.log.Error().Err(err).Msgf("Faield to read file %s", filePath)
					return
				}

				recordToInsert := sqlc.BulkInsertRecordParams{
					ID:            recordId,
					Userid:        record.UserId,
					Fileid:        fileId,
					Duration:      record.Duration,
					Distance:      record.Distance,
					Ascent:        record.Ascent,
					Descent:       record.Descent,
					Elevationdiff: record.ElevationDiff,
					Trails:        record.Trails,
					Rawdata:       string(fileData),
				}
				usersChan <- record.UserId
				recordChan <- recordToInsert
			}(record)
		}
		wg.Wait()

		close(recordChan)
		close(filesChan)
		close(usersChan)

		var dbwg sync.WaitGroup
		dbwg.Add(1)
		go func() {
			defer dbwg.Done()
			db.saveFilesToDatabase(preparedFiles)
			preparedFiles = nil
			db.saveRecordsToDatabase(preparedRecords)
			preparedRecords = nil
		}()
		dbwg.Wait()

        elapsedTime := time.Since(startTime)

        db.log.Info().Msgf("Completed batch %d / %d | Remaining: %d | Batch elapsed time: %v",
            i,
            batchCount,
            batchCount - (i + 1),
            elapsedTime,
        )
        db.log.Info().Msgf("Active goroutines: %d", runtime.NumGoroutine())
	}
}

func New(db *pgxpool.Pool, installPath string, log logger.Logger) Database {
	return &BaseDatabase{
		log:         log.With().Str("serivce", "database").Logger(),
		installPath: installPath,
		queries:     sqlc.New(db),
	}
}
