package services

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"time"

	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

/*
RunProcessorOnGroup reads the live lines data and the polygon data from the database.
It reads only the latest data for each collection. It does this by keeping track of the max inserted time for each collection.
If the max inserted time is less than the current time, it skips the collection.
If the max inserted time is missing or unable to read, it reads the entire collection and takes only the latest data.
It uses the live lines data to create the polygons.
Once it processes, it updates the polygons in the database and updates the live lines data in the database.
It then publishes the processed data to the rabbitmq queue with the group code as the routing key.
*/
func RunProcessorOnGroup(
	ctx context.Context,
	groupCode string,
	mongoClient *mongo.Client,
	amqpconfig *amqpconfig.AmqpConfig,
	log *zap.Logger,
	redisClient *redis.Client) error {

	currentMaxTime := getCurrentMaxTime(ctx, redisClient, groupCode, log)
	liveLinesData, err := getLiveLinesDataFromMongo(ctx, mongoClient, groupCode, log, currentMaxTime)
	if err != nil {
		return fmt.Errorf("Failed to get live lines data from mongo: %s", err)
	}

	polygonData, err := getPolygonDataFromMongo(ctx, mongoClient, groupCode, log, currentMaxTime)
	if err != nil {
		return fmt.Errorf("Failed to get polygon data from mongo: %s", err)
	}

	runProcessorConfig := RunProcessorConfig{
		DistanceFunc:            distance,
		NumSecondsForCycle:      10,
		DoSecondsForCycleCheck:  true,
		MetersThresholdForCycle: 1,
	}

	finalLiveLinesData, finalPolygonData, err := processData(liveLinesData, polygonData, runProcessorConfig)
	if err != nil {
		return fmt.Errorf("Failed to process data: %s", err)
	}

	currentMaxTime = time.Now().Nanosecond()
	for _, liveLines := range finalLiveLinesData {
		liveLines.InsertedTime = currentMaxTime
	}
	for _, polygonData := range finalPolygonData {
		polygonData.InsertedTime = currentMaxTime
	}

	// Add data into mongodb
	var wg sync.WaitGroup
	wg.Add(1)
	result := make(chan error)

	go func(result chan<- error) {
		defer wg.Done()
		session, err := mongoClient.StartSession()
		if err != nil {
			result <- fmt.Errorf("Failed to start session: %s", err)
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
			return putLiveLinesDataToMongo(ctx, mongoClient, groupCode, log, finalLiveLinesData)
		})

		if err != nil {
			result <- fmt.Errorf("Failed to put data to mongo: %s", err)
			return
		}

		_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
			return putPolygonDataToMongo(ctx, mongoClient, groupCode, log, finalPolygonData)
		})
		if err != nil {
			result <- fmt.Errorf("Failed to put data to mongo: %s", err)
			return
		}

		result <- nil
	}(result)

	wg.Wait()

	redisClient.Set(ctx, fmt.Sprintf("%s:%s", constants.REDIS_CURRENT_MAX_TIME, groupCode), currentMaxTime, 0)

	if err := <-result; err != nil {
		log.Error("Failed to put data to mongo", zap.Error(err))
		return fmt.Errorf("Failed to put data to mongo: %s", err)
	}

	finalData := dtoschema.RunnerResponse{
		Polygons:  finalPolygonData,
		LiveLines: finalLiveLinesData,
	}

	processedDataByte, err := json.Marshal(finalData)
	if err != nil {
		return fmt.Errorf("Failed to marshal data: %s", err)
	}

	amqpconfig.PublishWithContext([]byte(processedDataByte), groupCode, true)
	return nil

}

/*
processData checks if the live lines data for any user forms a cycle. If it does, it adds it to the polygons array.
Then it cleans the polygon array:
 1. Remove overlapping sections of older polygons.
 2. Merge polygons ran by the same user

Once it gets the cleaned polygons, it checks if any live lines points for each user should be removed.
It does so by checking the point inserted in the db is before the polygon was formed and if the point is inside the polygon.
It checks this for every polygon and every live line point.
*/
func processData(
	liveLinesData []dtoschema.LiveLinesData,
	polygonData []dtoschema.PolygonData,
	runProcessorConfig RunProcessorConfig) ([]dtoschema.LiveLinesData, []dtoschema.PolygonData, error) {
	// Get polygons that are formed by the live lines per user. Sort it be time in latest to oldest order

	// Step 1: Segregate the points
	mapOfLiveLinesForEachUSer := segregatePointsByUser(liveLinesData)

	// Step 2: Store users that form polygons in a separate map
	mapOfPolygonsForEachUser := createUserPolygonsMap(mapOfLiveLinesForEachUSer, runProcessorConfig)

	// Step 3: Add polygons to new polygon array. Add the existing polygons as well
	newPolygons := addPolygonsToNewPolygonArray(polygonData, mapOfPolygonsForEachUser)

	// Step 4: clean polygon data: Remove overlapping sections of older polygons. Merge polygons ran by the same user
	newPolygons = cleanPolygons(newPolygons)

	// TODO: Remove points from live lines data that are inside the polygons
	newLiveLinesData := removePointsFromLiveLinesData(liveLinesData, newPolygons)

	return newLiveLinesData, newPolygons, nil
}

/*
 * For each point in live lines data, check if it is inside any polygon.
 * If it is store the time of entry of the live lines data.
 * Any point by that runner which occurs before the time of entry is removed from the live lines data.
 * Do this for every runner.
 */
func removePointsFromLiveLinesData(liveLinesData []dtoschema.LiveLinesData, polygons []dtoschema.PolygonData) []dtoschema.LiveLinesData {
	mapLastPointBeforeDeletion := make(map[string]time.Time)

	for _, liveLinesData := range liveLinesData {

		runner := liveLinesData.Sender

		val, ok := mapLastPointBeforeDeletion[runner]
		if ok {
			if liveLinesData.Time.Before(val) {
				continue
			}
		}

		point := dtoschema.CoordinateStruct{
			X: liveLinesData.X,
			Y: liveLinesData.Y,
		}
		currMaxTime := liveLinesData.Time
		isAnyPointInPolygon := false

		for _, polygon := range polygons {
			temp := isPointInPolygon(point, polygon.Coords)
			if temp {
				isAnyPointInPolygon = true
				if polygon.Time.After(currMaxTime) {
					currMaxTime = polygon.Time
				}
			}
		}

		if isAnyPointInPolygon {
			mapLastPointBeforeDeletion[runner] = currMaxTime
		}
	}

	newLiveLinesData := []dtoschema.LiveLinesData{}

	for _, liveLinesData := range liveLinesData {
		runner := liveLinesData.Sender
		val, ok := mapLastPointBeforeDeletion[runner]
		if ok {
			if !liveLinesData.Time.After(val) {
				continue
			}
		}
		newLiveLinesData = append(newLiveLinesData, liveLinesData)
	}

	return newLiveLinesData
}

/*
cleanPolygons removes overlapping sections of older polygons.
Merge polygons ran by the same user.
*/
func cleanPolygons(polygons []dtoschema.PolygonData) []dtoschema.PolygonData {
	cleanedPolygons := []dtoschema.PolygonData{}
	slices.SortFunc(polygons, func(a, b dtoschema.PolygonData) int {
		return b.Time.Compare(a.Time)
	})

	for _, polygon := range polygons {
		for _, cleanedPolygon := range cleanedPolygons {
			// Merge polygons that belong to the same user later
			if polygon.Runner == cleanedPolygon.Runner {
				continue
			}
			polygon = eatPolygon(polygon, cleanedPolygon)
		}
		cleanedPolygons = append(cleanedPolygons, polygon)
	}

	cleanedPolygonsMerged := make(map[string][]dtoschema.PolygonData)

	for _, polygon := range cleanedPolygons {
		runner := polygon.Runner
		data, ok := cleanedPolygonsMerged[runner]
		if !ok {
			cleanedPolygonsMerged[runner] = append(cleanedPolygonsMerged[runner], polygon)
			continue
		}
		ok = false
		for idx, d := range data {
			newPolygon, canBeMerged := mergePolygons(polygon, d)
			ok = ok || canBeMerged
			if canBeMerged {
				data[idx] = newPolygon
				break
			}
		}
		if !ok {
			cleanedPolygonsMerged[runner] = append(cleanedPolygonsMerged[runner], polygon)
		}
	}

	cleanedPolygonsMergedList := []dtoschema.PolygonData{}
	runners := make([]string, 0, len(cleanedPolygonsMerged))
	for runner := range cleanedPolygonsMerged {
		runners = append(runners, runner)
	}
	slices.Sort(runners)

	for _, runner := range runners {
		polygon := cleanedPolygonsMerged[runner]
		for _, p := range polygon {
			cleanedPolygonsMergedList = append(cleanedPolygonsMergedList, p)
		}
	}

	return cleanedPolygonsMergedList
}

/*
Remove the points from p1 that are inside p2.
Add the points from p2 to p1.
Insert it in sorted order. Sorted by distance from the last deleted point
*/
func eatPolygon(p1 dtoschema.PolygonData, p2 dtoschema.PolygonData) dtoschema.PolygonData {
	pointsInP2, lastPointBeforeDeletion := getPointsInPolygon(p1.Coords, p2.Coords)
	pointsInP1, _ := getPointsInPolygon(p2.Coords, p1.Coords)

	if pointsInP2 == nil || pointsInP1 == nil {
		return p1
	}

	newPolygon := dtoschema.PolygonData{
		Runner: p1.Runner,
	}

	if lastPointBeforeDeletion != -1 {
		firstPoint := pointsInP1[0]
		lastPoint := pointsInP1[len(pointsInP1)-1]

		x := p1.Coords[lastPointBeforeDeletion].X
		y := p1.Coords[lastPointBeforeDeletion].Y

		if distance(firstPoint.X, firstPoint.Y, x, y) >
			distance(lastPoint.X, lastPoint.Y, x, y) {
			slices.Reverse(pointsInP1)
		}

		for idx, point := range p1.Coords {
			newPolygon.Coords = append(newPolygon.Coords, point)

			if idx == lastPointBeforeDeletion {
				for _, point := range pointsInP1 {
					newPolygon.Coords = append(newPolygon.Coords, point)
				}
			}
		}
	}

	// Remove points from p1 that are inside p2
	for _, point := range pointsInP2 {
		newPolygon.Coords = slices.DeleteFunc(newPolygon.Coords, func(p dtoschema.CoordinateStruct) bool {
			return point.X == p.X && point.Y == p.Y
		})
	}

	return newPolygon
}

/*
mergePolygons merges two polygons into one.
First it removes overlapping points from both input polygons.
Then it adds the points together from both polygons in sorted order.
*/
func mergePolygons(polygon1 dtoschema.PolygonData, polygon2 dtoschema.PolygonData) (dtoschema.PolygonData, bool) {
	// Get points from P1 that are inside P2
	pointsInPolygon2, _ := getPointsInPolygon(polygon1.Coords, polygon2.Coords)
	// Get points from P2 that are inside P1
	pointsInPolygon1, _ := getPointsInPolygon(polygon2.Coords, polygon1.Coords)

	if pointsInPolygon1 == nil || pointsInPolygon2 == nil {
		return polygon1, false
	}

	newPolygon := dtoschema.PolygonData{
		Runner: polygon1.Runner,
	}

	coords := make([]dtoschema.CoordinateStruct, 0)
	for _, point := range polygon1.Coords {
		coords = append(coords, point)
	}
	for _, point := range polygon2.Coords {
		coords = append(coords, point)
	}

	// Remove points from polygon1 that are inside polygon2
	for _, point := range pointsInPolygon2 {
		coords = slices.DeleteFunc(coords, func(p dtoschema.CoordinateStruct) bool {
			return point.X == p.X && point.Y == p.Y
		})
	}

	// Remove points from polygon2 that are inside polygon1
	for _, point := range pointsInPolygon1 {
		coords = slices.DeleteFunc(coords, func(p dtoschema.CoordinateStruct) bool {
			return point.X == p.X && point.Y == p.Y
		})
	}

	// Sort points by distance from each other in cyclic order. Start at the last point
	sortedCoords := sortPointsByDistance(coords)
	newPolygon.Coords = sortedCoords

	return newPolygon, true
}
