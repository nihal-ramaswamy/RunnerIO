package redisconfig

import (
	"context"

	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisConfig holds the configuration for the Redis client.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedisClient creates and initializes a new Redis client.
func NewRedisClient(ctx context.Context, log *zap.Logger, config *RedisConfig) *redis.Client {
	redisAddr := config.Host + ":" + config.Port
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Check the connection
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	log.Info("Connected to Redis")

	return client
}

// DefaultRedisConfig provides default Redis configuration.  It uses environment variables for flexibility.
func NewRedisConfig(options ...func(*RedisConfig)) *RedisConfig {
	config := &RedisConfig{
		Host:     utils.GetDotEnvVariable(constants.REDIS_HOST),
		Port:     utils.GetDotEnvVariable(constants.REDIS_PORT),
		Password: utils.GetDotEnvVariable(constants.REDIS_PASSWORD),
		DB:       0,
	}

	for _, option := range options {
		option(config)
	}

	return config
}

func WithRedisDB(db int) func(*RedisConfig) {
	return func(c *RedisConfig) {
		c.DB = db
	}
}

func DefaultRedisConfigForAuth() *RedisConfig {
	return NewRedisConfig(WithRedisDB(0))
}
