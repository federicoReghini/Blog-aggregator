package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(name string) error {
	config, err := Read()

	if err != nil {
		fmt.Println("Error while Reading config")
		return err
	}

	config.CurrentUserName = strings.Trim(name, " ")

	data, err := json.Marshal(config)

	if err != nil {

		fmt.Println("Error while Marshal")
		return err
	}

	write(data)
	return nil
}

func Read() (*Config, error) {
	// This function should read the configuration from a file or environment variables
	// and return a Config struct.

	path, err := getConfigPath()

	body, error := os.ReadFile(path)

	if error != nil {
		fmt.Println("Impossible to read file", path)
		return nil, err
	}

	var config Config

	if err := json.Unmarshal(body, &config); err != nil {
		fmt.Println("Impossible to deserialize json")
		return nil, err

	}

	return &config, nil

}

// Internal helpers function

func getConfigPath() (string, error) {

	home, err := os.UserHomeDir()

	if err != nil {
		fmt.Println("No home directory found")
		return "", err
	}

	return filepath.Join(home, configFileName), nil
}

func write(data []byte) {

	path, err := getConfigPath()

	if err != nil {
		fmt.Println("No home directory found")
	}

	os.WriteFile(path, data, 0644)
}
