package serverconfig

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
)

type Config struct {
	Port    string
	GinMode string // "debug", "release", "test"
	Cors    cors.Config
}

func NewServerConfig(options ...func(*Config)) *Config {
	config := &Config{}

	for _, option := range options {
		option(config)
	}

	return config
}

func WithPort(port string) func(*Config) {
	return func(c *Config) {
		c.Port = port
	}
}

func WithGinMode(ginMode string) func(*Config) {
	return func(c *Config) {
		switch ginMode {
		case gin.DebugMode:
			c.GinMode = gin.DebugMode
		case gin.ReleaseMode:
			c.GinMode = gin.ReleaseMode
		case gin.TestMode:
			c.GinMode = gin.TestMode
		default:
			log.Fatalf("Gin mode does not exist. Provided: %v", ginMode)
		}
	}
}

func WithCors(config cors.Config) func(*Config) {
	return func(c *Config) {
		c.Cors = config
	}
}

func WithCorsHosts(hosts []string) func(*Config) {
	return func(c *Config) {
		c.Cors.AllowOrigins = hosts
	}
}

func Default() *Config {
	return NewServerConfig(
		WithPort(utils.GetDotEnvVariable(constants.SERVER_PORT)),
		WithGinMode(utils.GetDotEnvVariable(constants.ENV)),
		WithCors(cors.DefaultConfig()),
		WithCorsHosts([]string{utils.GetDotEnvVariable(constants.SERVER_HOST) + utils.GetDotEnvVariable(constants.SERVER_PORT)}),
	)
}
