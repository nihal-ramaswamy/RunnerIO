package messages

import (
	"context"
	"math"

	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func getLiveLinesDataFromMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger) ([]dtoschema.LiveLinesData, error) {
	return getData[dtoschema.LiveLinesData](ctx, mongoClient, groupCode, log, constants.RUNNER_LIVE_LINES_COLLECTION)
}

func getPolygonDataFromMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger) ([]dtoschema.PolygonData, error) {
	return getData[dtoschema.PolygonData](ctx, mongoClient, groupCode, log, constants.RUNNER_POLYGON_COLLECTION)
}

func getData[T dtoschema.TestGenerics](ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger, collectionName string) ([]T, error) {
	cursor, err := mongoClient.Database(
		constants.RUNNER_DATABASE).Collection(
		collectionName).Find(
		ctx, bson.M{"group_code": groupCode})

	if err != nil {
		log.Error("Failed to find document", zap.Error(err))
		return nil, err
	}

	data := []T{}

	for cursor.Next(ctx) {
		var d bson.M
		if err := cursor.Decode(&d); err != nil {
			log.Error("Failed to decode document", zap.Error(err))
			continue
		}
		log.Info("Processing document", zap.Any("data", d))

		var groupData T
		if err := bson.UnmarshalExtJSON([]byte(d["data"].(string)), true, &groupData); err != nil {
			log.Error("Failed to unmarshal document", zap.Error(err))
			continue
		}

		data = append(data, groupData)
	}

	return data, nil
}

// https://gist.github.com/hotdang-ca/6c1ee75c48e515aec5bc6db6e3265e49
func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	radlat1 := float64(math.Pi * lat1 / 180)
	radlat2 := float64(math.Pi * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(math.Pi * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515

	dist = dist * 1.609344 * 1000

	return dist
}
