package downloader

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/Maxxxxxx-x/gpx-downloader/internal/logger"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/models"
)

const (
    DOWNLOAD_URL = "api"
	DOWNLOAD_BATCH_SIZE = 500
)

func downloadFile(fileName, outputPath string, log logger.Logger) error {
	if fileName == "" || outputPath == "" {
		return errors.New("Filename or OutputPath is empty")
	}

	filePath := path.Join(outputPath, fileName)
	outputFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		outputFile.Close()
		if err != nil {
			log.Error().Err(err).Msgf("ERROR OCCURED! REMOVING FILE %s", outputFile.Name())
			os.Remove(filePath)
		}
	}()

	log.Info().Msgf("Created file path at %s", filePath)

	resp, err := http.Get(fmt.Sprintf("%s%s", DOWNLOAD_URL, fileName))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return errors.New("Failed to download file. WE GOT RATE LIMITED!!!")
		}
		return fmt.Errorf("Failed to download file. Status code: %d", resp.StatusCode)
	}

	writtenBytes, err := io.Copy(outputFile, resp.Body)
	if err != nil {
		return err
	}

	log.Info().Msgf("Written %d bytes to %s", writtenBytes, outputFile.Name())

	return nil
}

func downloadBatch(installPath string, records []*models.DataRecord, size int, log logger.Logger) {
	skipped := 0
	fileCount := len(records)
	batchCount := int(math.Ceil(float64(fileCount / size)))

	for i := 0; i <= batchCount; i++ {
		lower := skipped
		upper := skipped + size
		if upper > fileCount {
			upper = fileCount
		}
		batchRecords := records[lower:upper]
		skipped += size

		errChan := make(chan error)
		doneChan := make(chan int)
		processErrors := make([]error, 0)

		go func() {
			for {
				select {
				case err := <-errChan:
					processErrors = append(processErrors, err)
				case <-doneChan:
					close(errChan)
					close(doneChan)
					return
				}
			}
		}()

		var wg sync.WaitGroup

		for _, record := range batchRecords{
            wg.Add(1)
			go func(instalLPath string, record *models.DataRecord) {
				defer wg.Done()
				if err := downloadFile(record.FileName, installPath, log); err != nil {
					errChan <- err
					log.Error().Err(err).Msgf("Failed to download %s", record.FileName)
				} else {
					log.Info().Msgf("Downloaded %s successfully", record.FileName)
				}
			}(installPath, record)
		}

		wg.Wait()

		doneChan <- 0
		log.Info().Msgf("Batch Download completed | Successful downloads: %d  | Errors occurred: %d", i, len(processErrors))
		if len(processErrors) > 0 {
			for _, err := range processErrors {
				log.Error().Err(err).Send()
			}
		}
	}
}

func StartDownload(csvRecords []*models.DataRecord, installPath string, log logger.Logger) {
	downloadBatch(installPath, csvRecords, DOWNLOAD_BATCH_SIZE, log)
}
