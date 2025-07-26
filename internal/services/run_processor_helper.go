package services

import (
	"context"
	"slices"

	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func putLiveLinesDataToMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger, liveLinesData []dtoschema.LiveLinesData) ([]interface{}, error) {
	return putData[dtoschema.LiveLinesData](ctx, mongoClient, groupCode, log, constants.RUNNER_LIVE_LINES_COLLECTION, liveLinesData)
}

func putPolygonDataToMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger, polygonData []dtoschema.PolygonData) ([]interface{}, error) {
	return putData[dtoschema.PolygonData](ctx, mongoClient, groupCode, log, constants.RUNNER_POLYGON_COLLECTION, polygonData)
}

func getLiveLinesDataFromMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger, currentMaxTime int) ([]dtoschema.LiveLinesData, error) {
	return getData[dtoschema.LiveLinesData](ctx, mongoClient, groupCode, log, constants.RUNNER_LIVE_LINES_COLLECTION, currentMaxTime)
}

func getPolygonDataFromMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger, currentMaxTime int) ([]dtoschema.PolygonData, error) {
	return getData[dtoschema.PolygonData](ctx, mongoClient, groupCode, log, constants.RUNNER_POLYGON_COLLECTION, currentMaxTime)
}

func segregatePointsByUser(liveLinesData []dtoschema.LiveLinesData) map[string][]dtoschema.LiveLinesData {
	mapOfLiveLinesForEachUSer := make(map[string][]dtoschema.LiveLinesData)
	for _, liveLine := range liveLinesData {
		mapOfLiveLinesForEachUSer[liveLine.Sender] = append(mapOfLiveLinesForEachUSer[liveLine.Sender], liveLine)
	}
	for userSub, liveLines := range mapOfLiveLinesForEachUSer {
		slices.SortFunc(liveLines, func(a, b dtoschema.LiveLinesData) int {
			return a.Time.Compare(b.Time)
		})
		mapOfLiveLinesForEachUSer[userSub] = liveLines
	}
	return mapOfLiveLinesForEachUSer
}

func createUserPolygonsMap(mapOfLiveLinesForEachUSer map[string][]dtoschema.LiveLinesData, runProcessorConfig RunProcessorConfig) map[string][][]dtoschema.LiveLinesData {
	mapOfPolygonsForEachUser := make(map[string][][]dtoschema.LiveLinesData)

	for userSub, liveLines := range mapOfLiveLinesForEachUSer {
		if !IsCycle(&liveLines, &runProcessorConfig) {
			continue
		}
		mapOfPolygonsForEachUser[userSub] = append(mapOfPolygonsForEachUser[userSub], liveLines)
	}
	return mapOfPolygonsForEachUser
}

func addPolygonsToNewPolygonArray(polygonData []dtoschema.PolygonData, mapOfPolygonsForEachUser map[string][][]dtoschema.LiveLinesData) []dtoschema.PolygonData {
	newPolygons := []dtoschema.PolygonData{}
	for _, polygon := range polygonData {
		newPolygons = append(newPolygons, polygon)
	}
	for _, polygons := range mapOfPolygonsForEachUser {
		for _, polygon := range polygons {
			newPolygons = append(newPolygons, dtoschema.ToPolygonData(&polygon))
		}
	}
	return newPolygons
}
