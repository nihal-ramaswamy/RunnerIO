package mongodbconfig

import (
	"context"

	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Config struct {
	UserName string
	Password string
	Host     string
	Port     string
}

func NewConfig(options ...func(*Config)) *Config {
	config := &Config{}

	for _, option := range options {
		option(config)
	}
	return config
}

func WithUserName(userName string) func(*Config) {
	return func(c *Config) {
		c.UserName = userName
	}
}
func WithPassword(password string) func(*Config) {
	return func(c *Config) {
		c.Password = password
	}
}

func WithHost(host string) func(*Config) {
	return func(c *Config) {
		c.Host = host
	}
}

func WithPort(port string) func(*Config) {
	return func(c *Config) {
		c.Port = port
	}
}

func DefaultConfig() *Config {
	return NewConfig(
		WithUserName(utils.GetDotEnvVariable(constants.MONGO_INITDB_ROOT_USERNAME)),
		WithPassword(utils.GetDotEnvVariable(constants.MONGO_INITDB_ROOT_PASSWORD)),
		WithHost(utils.GetDotEnvVariable(constants.MONGO_HOST)),
		WithPort(utils.GetDotEnvVariable(constants.ME_CONFIG_MONGODB_PORT)),
	)
}

func (c *Config) GetURI() string {
	// mongodb://username:password@host:port/
	return "mongodb://" + c.UserName + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/"
}

func Connect(config *Config, log *zap.Logger) *mongo.Client {
	MONGO_URI := config.GetURI()

	// Connect to the database.
	clientOption := options.Client().ApplyURI(MONGO_URI)
	client, err := mongo.Connect(context.Background(), clientOption)
	if err != nil {
		log.Fatal("Error connecting to MongoDB", zap.Error(err))
	}

	// Check the connection.
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB", zap.Error(err), zap.String("uri", MONGO_URI))
	}

	log.Info("Connected to MongoDB")

	return client
}
