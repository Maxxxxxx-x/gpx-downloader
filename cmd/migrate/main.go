package main

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	c "github.com/Maxxxxxx-x/gpx-downloader/internal/config"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/logger"
	"github.com/Maxxxxxx-x/gpx-downloader/internal/sql/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
)

func main() {
	config, err := c.GetConfig("migration")
	if err != nil {
		fmt.Printf("%v", err)
	}
	config.Logging.LogPath = "./logs/migration/migrate.log"
	log := logger.New(config.Logging, config.Env)
	if err != nil {
		log.Fatal().Err(err).Msg("Error occured while loading config. Migration failed")
	}

	log.Info().Msg("Config loaded. Starting migration...")

	db, err := sql.Open("pgx", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host,
		config.Database.Port,
		config.Database.Username,
		config.Database.Password,
		config.Database.DatabaseName,
	))

	if err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to db")
	}
	log.Info().Msg("Connected to database")
	defer db.Close()
	provider, err := goose.NewProvider(database.DialectPostgres, db, migrations.Embed)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create goose provider")
	}

	log.Info().Msg("====== migration list ======")
	sources := provider.ListSources()
	for _, source := range sources {
		log.Info().Msgf("%-3s %-2v %v", source.Type, source.Version, filepath.Base(source.Path))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	status, err := provider.Status(ctx)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Info().Msg("====== Migration Status ======")
	for _, stats := range status {
		log.Info().Msgf("%-3s %-2v %v", stats.Source.Type, stats.Source.Version, stats.State)
	}

	log.Info().Msg("====== Migration Logs ======")
	results, err := provider.Up(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to apply migration")
	}
	log.Info().Msg("====== Migration Results ======")
	for _, result := range results {
		log.Info().Msgf("%-3v %-2v done: %v", result.Source.Type, result.Source.Version, result.Duration)
	}

}
