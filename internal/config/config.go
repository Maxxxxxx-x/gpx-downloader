package config

import (
	"fmt"
	"os"
	"path"

	"github.com/Maxxxxxx-x/gpx-downloader/internal/utils"
	"gopkg.in/yaml.v3"
)

type LoggingConfig struct {
	LogPath      string `yaml:"LogPath"`
	LogLevel     string `yaml:"LogLevel"`
	CompressLogs bool   `yaml:"CompressLogs"`
	MaxSize      int    `yaml:"MaxSize"`
	MaxAge       int    `yaml:"MaxAge"`
	MaxBackups   int    `yaml:"MaxBackup"`
}

type DatabaseConfig struct {
	Enabled      bool   `yaml:"Enabled"`
	Host         string `yaml:"Host"`
	Port         string `yaml:"Port"`
	Username     string `yaml:"Username"`
	PasswordPath string `yaml:"PasswordPath"`
	Password     string
	DatabaseName string `yaml:"DBName"`
}

type Config struct {
	Env      string
	Database DatabaseConfig `yaml:"Database"`
	Logging  LoggingConfig  `yaml:"Logging"`
}

func missingEnv(envName string) error {
	return fmt.Errorf("Missing ENV variable: %s\n", envName)
}

func readFile(filePath string) ([]byte, error) {
	if _, err := os.Stat(filePath); err != nil {
		return []byte{}, err
	}
	return os.ReadFile(filePath)
}

func getConfigFile(fileName, env string) ([]byte, error) {
	configPath := path.Join("/config", fmt.Sprintf("%s.yml", fileName))
	if env == "dev" {
		execPath, err := os.Executable()
		if err != nil {
			return []byte{}, err
		}
		configPath = path.Join(path.Dir(execPath), fmt.Sprintf("../../config/%s.yml", fileName))
	}
	return readFile(configPath)
}

func loadDefaultLogging() LoggingConfig {
	return LoggingConfig{}
}

func GetConfig(fileName string) (Config, error) {
	config := Config{}
	env, found := os.LookupEnv("APP_ENV")
	if !found {
		env = "dev"
	}

	config.Env = env

	configBytes, err := getConfigFile(fileName, config.Env)
	if err != nil {
		config.Logging = loadDefaultLogging()
	}

	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return config, err
	}

    if config.Database.Enabled {
        var passwordBytes []byte
        if env == "prod" {
            passwordBytes, err = readFile(config.Database.PasswordPath)
            if err != nil {
                return config, err
            }
        } else {
            execPath, err := os.Executable()
            if err != nil {
                return config, err
            }
            pwdPath := path.Join(path.Dir(execPath), fmt.Sprintf("../../secrets/%s_password", fileName))
            passwordBytes, err = readFile(pwdPath)
            if err != nil {
                return config, err
            }
        }
        config.Database.Password = string(passwordBytes)
    }

	if valid, err := utils.ValidatePort(config.Database.Port); err != nil || !valid {
		return config, err
	}

	return config, nil
}
