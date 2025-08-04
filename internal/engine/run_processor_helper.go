package services

import (
	"context"
	"slices"

	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/mongodb_schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func putLiveLinesDataToMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger, liveLinesData []dtoschema.RunnerLiveLinesData) ([]interface{}, error) {
	return putData[dtoschema.RunnerLiveLinesData](ctx, mongoClient, groupCode, log, constants.RUNNER_LIVE_LINES_COLLECTION, liveLinesData)
}

func putPolygonDataToMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger, polygonData []dtoschema.RunnerPolygonSchema) ([]interface{}, error) {
	return putData[dtoschema.RunnerPolygonSchema](ctx, mongoClient, groupCode, log, constants.RUNNER_POLYGON_COLLECTION, polygonData)
}

func getLiveLinesDataFromMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger, currentMaxTime int) ([]dtoschema.RunnerLiveLinesData, error) {
	return getData[dtoschema.RunnerLiveLinesData](ctx, mongoClient, groupCode, log, constants.RUNNER_LIVE_LINES_COLLECTION, currentMaxTime)
}

func getPolygonDataFromMongo(ctx context.Context, mongoClient *mongo.Client, groupCode string, log *zap.Logger, currentMaxTime int) ([]dtoschema.RunnerPolygonSchema, error) {
	return getData[dtoschema.RunnerPolygonSchema](ctx, mongoClient, groupCode, log, constants.RUNNER_POLYGON_COLLECTION, currentMaxTime)
}

func segregatePointsByUser(liveLinesData []dtoschema.RunnerLiveLinesData) map[string][]dtoschema.RunnerLiveLinesData {
	mapOfLiveLinesForEachUSer := make(map[string][]dtoschema.RunnerLiveLinesData)
	for _, liveLine := range liveLinesData {
		mapOfLiveLinesForEachUSer[liveLine.Sender] = append(mapOfLiveLinesForEachUSer[liveLine.Sender], liveLine)
	}
	for userSub, liveLines := range mapOfLiveLinesForEachUSer {
		slices.SortFunc(liveLines, func(a, b dtoschema.RunnerLiveLinesData) int {
			return a.Time.Compare(b.Time)
		})
		mapOfLiveLinesForEachUSer[userSub] = liveLines
	}
	return mapOfLiveLinesForEachUSer
}

func createUserPolygonsMap(mapOfLiveLinesForEachUSer map[string][]dtoschema.RunnerLiveLinesData, runProcessorConfig RunProcessorConfig) map[string][][]dtoschema.RunnerLiveLinesData {
	mapOfPolygonsForEachUser := make(map[string][][]dtoschema.RunnerLiveLinesData)

	for userSub, liveLines := range mapOfLiveLinesForEachUSer {
		if !IsCycle(&liveLines, &runProcessorConfig) {
			continue
		}
		mapOfPolygonsForEachUser[userSub] = append(mapOfPolygonsForEachUser[userSub], liveLines)
	}
	return mapOfPolygonsForEachUser
}

func addPolygonsToNewPolygonArray(polygonData []dtoschema.RunnerPolygonSchema, mapOfPolygonsForEachUser map[string][][]dtoschema.RunnerLiveLinesData) []dtoschema.RunnerPolygonSchema {
	newPolygons := []dtoschema.RunnerPolygonSchema{}
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
