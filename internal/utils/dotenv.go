package utils

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
)

// TODO: Clean this piece of code.
func validateEnvVariable(value string) (bool, error) {
	if value == "debug" || value == "release" || value == "test" {
		return true, nil
	}

	log.Fatalf("ENV value must be 'debug' or 'release' or 'test'. Provided: %v", value)
	return false, errors.New("Invalid value set for ENV")
}

func GetDotEnvVariable(key string) string {

	err := godotenv.Load()

	if nil != err {
		log.Fatalf("Error loading .env file")
	}

	value := os.Getenv(key)

	if key == constants.ENV {
		if _, err := validateEnvVariable(value); nil != err {
			os.Exit(1)
		}
	}

	return value
}
