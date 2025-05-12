package constants

import "time"

const (
	// Timme for tooke to expire
	TOKEN_EXPIRY_TIME = time.Hour * 24 * 7 // 7 days

	POSTGRES_HOST     = "POSTGRES_HOST"
	POSTGRES_PORT     = "POSTGRES_PORT"
	POSTGRES_USER     = "POSTGRES_USER"
	POSTGRES_PASSWORD = "POSTGRES_PASSWORD"
	POSTGRES_NAME     = "POSTGRES_NAME"

	REDIS_PASSWORD = "REDIS_PASSWORD"
	REDIS_HOST     = "REDIS_HOST"
	REDIS_PORT     = "REDIS_PORT"

	SERVER_PORT = "SERVER_PORT"
	SERVER_HOST = "SERVER_HOST"

	SECRET_KEY = "SECRET_KEY"

	ENV = "ENV"

	// UUID generation
	UUID_CHARACTERS = "qwertyuiopasdfghjklzxcvbnm"
	UUID_LENGTH     = 9
)

func GetRuneUuidCharacters() []rune {
	return []rune(UUID_CHARACTERS)
}
