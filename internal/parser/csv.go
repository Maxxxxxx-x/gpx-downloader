package parser

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Maxxxxxx-x/gpx-downloader/internal/logger"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/models"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/utils"
	"github.com/gocarina/gocsv"
)

type ParserResult struct {
	File  models.CSVFile
	Error error
}

func getFilesFromDirectory(dirPath string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return files, err
	}

	exc, err := os.Executable()
	if err != nil {
		return files, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".csv" {
			continue
		}
		files = append(files, path.Join(path.Dir(exc), "..", "..", dirPath, entry.Name()))
	}

	return files, nil
}

func errorOccured(fileName string, err error) error {
	return fmt.Errorf("Error occurred while attempting to parse %s. Error: %s", fileName, err.Error())
}

func parseCSVFile(filePath string, resChan chan ParserResult, log logger.Logger, wg *sync.WaitGroup) {
	var content models.CSVFile
	defer wg.Done()

	log.Info().Msgf("Opening %s", filePath)
	file, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("Erorr occured opening file")
		filePathParts := strings.Split(filePath, "/")
		fileName := filePathParts[len(filePathParts)-1]
		resChan <- ParserResult{
			File: models.CSVFile{
				FileName: fileName,
			},
			Error: errorOccured(fileName, err),
		}
		return
	}
	defer file.Close()
	content.FileName = file.Name()

	if err := gocsv.UnmarshalFile(file, &content.Data); err != nil {
		log.Error().Err(err).Msgf("Erorr occured parsing CSV file %s", file.Name())
		resChan <- ParserResult{
			File:  content,
			Error: errorOccured(file.Name(), err),
		}
		return
	}

    fileHash, err := utils.GenerateFileHash(file)
    if err != nil {
        log.Error().Err(err).Msgf("Error occured generating file hash for file %s", file.Name())
        resChan <- ParserResult{
            File:  content,
            Error: errorOccured(file.Name(), err),
        }
        return
    }
    content.SHA512Sum = fileHash

    resChan <- ParserResult{
        File:  content,
        Error: nil,
    }

	log.Info().Msgf("Parssed %s successfully!", filePath)
	return
}

func StartParser(sourcePath string, log logger.Logger) ([]models.CSVFile, []error) {
	var csvFiles []models.CSVFile
	var errors []error

	startTime := time.Now()
	logger := log.With().Str("service", "CSV Parser").Logger()
	log.Info().Msgf("Getting CSV files from %s", sourcePath)
	files, err := getFilesFromDirectory(sourcePath)
	if err != nil {
		return csvFiles, []error{err}
	}

	log.Info().Msgf("%d CSV files found. Starting CSV parsing...", len(files))
	wg := sync.WaitGroup{}
	resChan := make(chan ParserResult)

	defer utils.Timer("CSV Parser", logger)()

	for _, filePath := range files {
		wg.Add(1)
		go parseCSVFile(filePath, resChan, log, &wg)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for res := range resChan {
		csvFiles = append(csvFiles, res.File)
		errors = append(errors, res.Error)
	}

	totalErr := 0
	for _, err := range errors {
		if err != nil {
			totalErr += 1
			log.Error().Err(err).Send()
		}
	}

	log.Info().Msgf(
		"CSV Parser completed! | Files Parsed: %d | Errors: %d | Elapsed Time: %v",
		len(csvFiles),
        totalErr,
		time.Since(startTime),
	)

	return csvFiles, errors
}
