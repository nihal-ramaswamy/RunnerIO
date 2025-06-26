package redisconfig

import (
	"context"
	"time"

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
func NewRedisClient(ctx context.Context, log *zap.Logger, config RedisConfig) *redis.Client {
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

	return client
}

// DefaultRedisConfig provides default Redis configuration.  It uses environment variables for flexibility.
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     utils.GetDotEnvVariable(constants.REDIS_HOST),
		Port:     utils.GetDotEnvVariable(constants.REDIS_PORT),
		Password: utils.GetDotEnvVariable(constants.REDIS_PASSWORD),
		DB:       1,
	}
}

// WithRedisConnectionTimeout sets a connection timeout for the Redis client.
func WithRedisConnectionTimeout(timeout time.Duration) func(*redis.Options) {
	return func(opts *redis.Options) {
		opts.DialTimeout = timeout
	}
}
