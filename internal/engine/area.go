package services

import (
	"context"
	"math"

	mongo_schema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/mongodb_schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func AreaProcessor(ctx context.Context, groupCode string, mongoClient *mongo.Client, log *zap.Logger, currentMaxTime int) (map[string]float64, error) {
	areaForEachUser := make(map[string]float64)

	polygonData, err := getPolygonDataFromMongo(ctx, mongoClient, groupCode, log, currentMaxTime)
	if err != nil {
		log.Error("Failed to get polygon data from mongo", zap.Error(err))
		return nil, err
	}

	for _, polygon := range polygonData {
		area := getPolygonArea(polygon.Coords)
		currentArea, ok := areaForEachUser[polygon.Runner]
		if ok {
			areaForEachUser[polygon.Runner] = currentArea + area
			continue
		}
		areaForEachUser[polygon.Runner] = area
	}

	return areaForEachUser, nil
}

// Shoelace formula
func getPolygonArea(coords []mongo_schema.CoordinateStruct) float64 {
	coords = sortPointsByDistance(coords)
	area := 0.0

	for i := 0; i < len(coords); i++ {
		x1 := coords[i].X
		y1 := coords[i].Y
		x2 := coords[(i+1)%len(coords)].X
		y2 := coords[(i+1)%len(coords)].Y

		area += (x1*y2 - x2*y1)
	}

	return math.Abs(area / 2)
}
