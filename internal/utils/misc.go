package utils

import (
	"context"
	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func CreateCode(mongoClient *mongo.Client, log *zap.Logger) string {
	var result dtoschema.RunnerGroupSchema

	for {
		code := GenerateRandomString(10)
		filter := bson.M{"code": code}
		err := mongoClient.Database(constants.RUNNER_DATABASE).Collection(constants.RUNNER_GROUP_COLLECTION).FindOne(context.TODO(), filter).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return code
			}
			log.Error("Error generating group code.", zap.Error(err))
		}
	}
}

func GenerateRandomString(length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func FailIfError(err error, c *gin.Context, log *zap.Logger, code int, message string, fields ...zap.Field) {
	if err == nil {
		return
	}

	log.Error(message, fields...)
	c.AbortWithError(code, err)
}
