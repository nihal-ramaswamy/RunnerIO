package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

/*
RunProcessorOnGroup reads the live lines data from the database.
It uses the live lines data to create the polygons.
Once it processes, it updates the polygons in the database and updates the live lines data in the database.
It then publishes the processed data to the rabbitmq queue with the group code as the routing key.
*/
func RunProcessorOnGroup(
	ctx context.Context,
	groupCode string,
	mongoClient *mongo.Client,
	amqpconfig *amqpconfig.AmqpConfig,
	log *zap.Logger) error {

	liveLinesData, err := getLiveLinesDataFromMongo(ctx, mongoClient, groupCode, log)
	if err != nil {
		return fmt.Errorf("Failed to get live lines data from mongo: %s", err)
	}

	polygonData, err := getPolygonDataFromMongo(ctx, mongoClient, groupCode, log)
	if err != nil {
		return fmt.Errorf("Failed to get polygon data from mongo: %s", err)
	}

	runProcessorConfig := RunProcessorConfig{
		DistanceFunc:       distance,
		NumSecondsForCycle: 10,
	}

	finalData, err := processData(liveLinesData, polygonData, runProcessorConfig)
	if err != nil {
		return fmt.Errorf("Failed to process data: %s", err)
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
	runProcessorConfig RunProcessorConfig) ([]dtoschema.PolygonData, error) {
	// Get polygons that are formed by the live lines per user. Sort it be time in latest to oldest order

	// Step 1: Segregate the points
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

	// Step 2: Store users that form polygons in a separate map
	mapOfPolygonsForEachUser := make(map[string][][]dtoschema.LiveLinesData)

	for userSub, liveLines := range mapOfLiveLinesForEachUSer {
		if !IsCycle(liveLines, runProcessorConfig) {
			continue
		}
		mapOfPolygonsForEachUser[userSub] = append(mapOfPolygonsForEachUser[userSub], liveLines)
	}

	// Step 3: Add polygons to new polygon array. Add the existing polygons as well
	newPolygons := []dtoschema.PolygonData{}
	for _, polygon := range polygonData {
		newPolygons = append(newPolygons, polygon)
	}
	for _, polygons := range mapOfPolygonsForEachUser {
		for _, polygon := range polygons {
			newPolygons = append(newPolygons, *dtoschema.ToPolygonData(&polygon))
		}
	}

	// Step 4: clean polygon data: Remove overlapping sections of older polygons. Merge polygons ran by the same user
	newPolygons = cleanPolygons(newPolygons)

	// TOOD: Remove points from live lines data that are inside the polygons

	return newPolygons, nil
}

/*
isCycle checks if the live lines data forms a cycle.

A cycle is formed when the first and last point of the live lines data are within a distance of half a meter and the time difference is more than 10 seconds.
*/
func IsCycle(liveLines []dtoschema.LiveLinesData, runProcessorConfig RunProcessorConfig) bool {
	if len(liveLines) < 2 {
		return false
	}

	firstPoint := liveLines[0]
	lastPoint := liveLines[len(liveLines)-1]

	// Rule for a cycle: The first and last point should have a distance of less than half a meter and the time difference should be more than 10 seconds
	if firstPoint.Time.Sub(lastPoint.Time).Seconds() < float64(runProcessorConfig.NumSecondsForCycle)-constants.DELTA {
		return false
	}

	if runProcessorConfig.DistanceFunc(firstPoint.X, firstPoint.Y, lastPoint.X, lastPoint.Y) >
		runProcessorConfig.MetersThresholdForCycle-constants.DELTA {
		return false
	}
	return true
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
	for _, polygon := range cleanedPolygonsMerged {
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

// Get points from P1 that are inside P2.
// Also returns the index of the point before which the overlapping points were found
func getPointsInPolygon(P1 []dtoschema.CoordinateStruct, P2 []dtoschema.CoordinateStruct) ([]dtoschema.CoordinateStruct, int) {
	var pointsInPolygon []dtoschema.CoordinateStruct
	pointsIdxInPolygon := []int{}

	for idx, point := range P1 {
		if isPointInPolygon(point, P2) {
			pointsInPolygon = append(pointsInPolygon, point)
			pointsIdxInPolygon = append(pointsIdxInPolygon, idx)
		}
	}

	if len(pointsIdxInPolygon) == 0 {
		return nil, -1
	}

	if len(pointsIdxInPolygon) == len(P1) {
		return pointsInPolygon, -1
	}

	lastPointIdxBeforeDeletion := (pointsIdxInPolygon[0] - 1 + len(P1)) % len(P1)
	for slices.Contains(pointsIdxInPolygon, lastPointIdxBeforeDeletion) {
		lastPointIdxBeforeDeletion -= 1
		if lastPointIdxBeforeDeletion == -1 {
			lastPointIdxBeforeDeletion = len(P1) - 1
		}
	}

	slices.SortFunc(pointsInPolygon, func(a, b dtoschema.CoordinateStruct) int {
		tempA := a.X*a.X + a.Y*a.Y
		tempB := b.X*b.X + b.Y*b.Y
		return int(tempA - tempB)
	})

	return pointsInPolygon, lastPointIdxBeforeDeletion
}

// Function checks if point is inside the polygon
// Check is done using ray casting algorithm
func isPointInPolygon(point dtoschema.CoordinateStruct, polygon []dtoschema.CoordinateStruct) bool {
	onside := slices.ContainsFunc(polygon, func(p dtoschema.CoordinateStruct) bool {
		return p.X == point.X && p.Y == point.Y
	})
	if onside {
		return true
	}

	var inside bool
	for i, p := range polygon {
		j := i + 1
		if j == len(polygon) {
			j = 0
		}
		pj := polygon[j]
		if ((p.Y > point.Y) != (pj.Y > point.Y)) &&
			(point.X < (pj.X-p.X)*(point.Y-p.Y)/(pj.Y-p.Y)+p.X) {
			inside = !inside
		}
	}
	return inside
}

// Start at a random point and find the next point closest to this point. Update the start point with the new point
// Repeat until all points are sorted
func sortPointsByDistance(points []dtoschema.CoordinateStruct) []dtoschema.CoordinateStruct {
	if len(points) <= 2 {
		return points
	}

	startPoint := points[0]
	sortedPoints := []dtoschema.CoordinateStruct{startPoint}
	visited := make([]bool, len(points))
	visited[0] = true

	for {
		minDistance := 1e9 + 7
		p := -1
		for i, point := range points {
			if visited[i] {
				continue
			}

			startPoint = sortedPoints[len(sortedPoints)-1]

			distance := distance(startPoint.X, startPoint.Y, point.X, point.Y)
			if distance < minDistance {
				minDistance = distance
				p = i
			}
		}
		if p == -1 {
			break
		}
		sortedPoints = append(sortedPoints, points[p])
		visited[p] = true
	}
	return sortedPoints
}
