package utils

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
)

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
		log.Panicf("Error loading .env file: %v", err)
	}

	value := os.Getenv(key)

	if key == constants.ENV {
		if _, err := validateEnvVariable(value); nil != err {
			os.Exit(1)
		}
	}

	if key == constants.SERVER_PORT {
		val := os.Getenv(constants.SERVER)
		if val != constants.SERVICE && val != constants.ENGINE {
			log.Fatalf("Server must be either 'service' or 'engine'. Provided: %v", val)
		}
		switch val {
		case constants.SERVICE:
			value = os.Getenv(constants.SERVER_SERVICE_PORT)
		case constants.ENGINE:
			value = os.Getenv(constants.SERVER_ENGINE_PORT)
		}
	}

	return value
}
