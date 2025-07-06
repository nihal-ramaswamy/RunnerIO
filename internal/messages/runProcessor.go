package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
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

	finalData, err := processData(&liveLinesData, &polygonData)
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
func processData(liveLinesData *[]dtoschema.RunnerLiveLinesSchema, polygonData *[]dtoschema.PolygonData) ([]dtoschema.RunnerEatAreaSchema, error) {
	// Get polygons that are formed by the live lines per user. Sort it be time in latest to oldest order

	// Step 1: Segregate the points
	mapOfLiveLinesForEachUSer := make(map[string][]dtoschema.RunnerLiveLinesSchema)
	for _, liveLine := range *liveLinesData {
		mapOfLiveLinesForEachUSer[liveLine.Sender] = append(mapOfLiveLinesForEachUSer[liveLine.Sender], liveLine)
	}
	for userSub, liveLines := range mapOfLiveLinesForEachUSer {
		slices.SortFunc(liveLines, func(a, b dtoschema.RunnerLiveLinesSchema) int {
			return a.Time.Compare(b.Time)
		})
		mapOfLiveLinesForEachUSer[userSub] = liveLines
	}

	// Step 2: Store users that form polygons in a separate map
	mapOfPolygonsForEachUser := make(map[string][][]dtoschema.RunnerLiveLinesSchema)

	for userSub, liveLines := range mapOfLiveLinesForEachUSer {
		if !isCycle(liveLines) {
			continue
		}
		mapOfPolygonsForEachUser[userSub] = append(mapOfPolygonsForEachUser[userSub], liveLines)
	}

	// Step 3: Add polygons to new polygon array. Add the existing polygons as well
	newPolygons := []dtoschema.PolygonData{}
	for _, polygon := range *polygonData {
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

	return []dtoschema.RunnerEatAreaSchema{}, nil
}

/*
isCycle checks if the live lines data forms a cycle.

A cycle is formed when the first and last point of the live lines data are within a distance of half a meter and the time difference is more than 10 seconds.
*/
func isCycle(liveLines []dtoschema.RunnerLiveLinesSchema) bool {
	if len(liveLines) < 2 {
		return false
	}

	firstPoint := liveLines[0]
	lastPoint := liveLines[len(liveLines)-1]

	// Rule for a cycle: The first and last point should have a distance of less than half a meter and the time difference should be more than 10 seconds
	if firstPoint.Time.Sub(lastPoint.Time).Seconds() < 10 {
		return false
	}
	if distance(firstPoint.X, firstPoint.Y, lastPoint.X, lastPoint.Y) > 0.5+1e-8 {
		return false
	}
	return true
}

/*
cleanPolygons removes overlapping sections of older polygons. Merge polygons ran by the same user.
*/
func cleanPolygons(polygons []dtoschema.PolygonData) []dtoschema.PolygonData {
	cleanedPolygons := []dtoschema.PolygonData{}
	slices.SortFunc(polygons, func(a, b dtoschema.PolygonData) int {
		return b.Time.Compare(a.Time)
	})

	// Flood fill algorithm to remove overlapping sections
	cleanedPolygons = append(cleanedPolygons, polygons[0])

	for i := 1; i < len(polygons); i++ {
		polygon := polygons[i]
		for j := 0; j < len(cleanedPolygons); j++ {
			cleanedPolygon := cleanedPolygons[j]
			// Merge polygons that belong to the same user later
			if polygon.Runner == cleanedPolygon.Runner {
				continue
			}
			polygon = eatPolygon(polygon, cleanedPolygon)
		}
		cleanedPolygons = append(cleanedPolygons, polygon)
	}

	cleanedPolygonsMerged := make(map[string]dtoschema.PolygonData)

	for i := 0; i < len(cleanedPolygons); i++ {
		runner := cleanedPolygons[i].Runner
		data, ok := cleanedPolygonsMerged[runner]
		if !ok {
			cleanedPolygonsMerged[runner] = cleanedPolygons[i]
			continue
		}
		cleanedPolygonsMerged[runner] = mergePolygons(cleanedPolygons[i], data)
	}

	cleanedPolygonsMergedList := []dtoschema.PolygonData{}
	for _, polygon := range cleanedPolygonsMerged {
		cleanedPolygonsMergedList = append(cleanedPolygonsMergedList, polygon)
	}

	return cleanedPolygonsMergedList
}

/*
Remove the points from polygon that are inside cleanedPolygon.
Add the points from cleanedPolygon to polygon.
Insert it in sorted order. Sorted by distance from the last deleted point
*/
func eatPolygon(polygon dtoschema.PolygonData, cleanedPolygon dtoschema.PolygonData) dtoschema.PolygonData {
	pointsInPolygon, lastPointBeforeDeletion := getPointsInPolygon(cleanedPolygon.Coords, polygon.Coords)
	pointsInCleanedPolygon, _ := getPointsInPolygon(polygon.Coords, cleanedPolygon.Coords)

	if pointsInCleanedPolygon == nil || pointsInPolygon == nil {
		return polygon
	}

	// Remove points from polygon that are inside cleanedPolygon
	for _, point := range pointsInCleanedPolygon {
		polygon.Coords = slices.DeleteFunc(polygon.Coords, func(p dtoschema.CoordinateStruct) bool {
			return point.X == p.X && point.Y == p.Y
		})
	}

	if lastPointBeforeDeletion == nil {
		return polygon
	}

	// Add points from cleanedPolygon to polygon. Insert it in sorted order. Sorted by distance from the last deleted point
	idxOfLastPointBeforeDeletion := -1
	for idx, point := range polygon.Coords {
		if point.X == lastPointBeforeDeletion.X && point.Y == lastPointBeforeDeletion.Y {
			idxOfLastPointBeforeDeletion = idx
			break
		}
	}
	firstPoint := pointsInPolygon[0]
	lastPoint := pointsInPolygon[len(pointsInPolygon)-1]

	if distance(firstPoint.X, firstPoint.Y, lastPointBeforeDeletion.X, lastPointBeforeDeletion.Y) >
		distance(lastPoint.X, lastPoint.Y, lastPointBeforeDeletion.X, lastPointBeforeDeletion.Y) {
		slices.Reverse(pointsInPolygon)
	}
	// Insert points from pointsInPolygon into polygon starting from idxOfLastPointBeforeDeletion
	polygon.Coords = append(polygon.Coords[:idxOfLastPointBeforeDeletion], pointsInPolygon...)

	return polygon
}

/*
mergePolygons merges two polygons.
It gets the points from P1 that are inside P2.
Then it removes the points from polygon1 that are inside polygon2.
Then it adds the points from polygon2 to polygon1.
Insert it in sorted order. Sorted by distance from the last deleted point of P1.
It uses the fact that the points are already sorted by distance (or time)
*/
func mergePolygons(polygon1 dtoschema.PolygonData, polygon2 dtoschema.PolygonData) dtoschema.PolygonData {
	// Get points from P1 that are inside P2
	pointsInPolygon2, lastPointBeforeDeletionP1 := getPointsInPolygon(polygon1.Coords, polygon2.Coords)
	pointsInPolygon1, _ := getPointsInPolygon(polygon2.Coords, polygon1.Coords)

	if pointsInPolygon1 == nil || pointsInPolygon2 == nil {
		return polygon1
	}

	// Remove points from polygon1 that are inside polygon2
	for _, point := range pointsInPolygon2 {
		polygon1.Coords = slices.DeleteFunc(polygon1.Coords, func(p dtoschema.CoordinateStruct) bool {
			return point.X == p.X && point.Y == p.Y
		})
	}

	if lastPointBeforeDeletionP1 == nil {
		return polygon1
	}

	// Add points from polygon2 to polygon1. Insert it in sorted order. Sorted by distance from the last deleted point
	idxOfLastPointBeforeDeletionP1 := -1
	for idx, point := range polygon1.Coords {
		if point.X == lastPointBeforeDeletionP1.X && point.Y == lastPointBeforeDeletionP1.Y {
			idxOfLastPointBeforeDeletionP1 = idx
			break
		}
	}
	firstPointP1 := pointsInPolygon1[0]
	lastPointP1 := pointsInPolygon1[len(pointsInPolygon1)-1]

	if distance(firstPointP1.X, firstPointP1.Y, lastPointBeforeDeletionP1.X, lastPointBeforeDeletionP1.Y) >
		distance(lastPointP1.X, lastPointP1.Y, lastPointBeforeDeletionP1.X, lastPointBeforeDeletionP1.Y) {
		slices.Reverse(pointsInPolygon1)
	}
	// Insert points from pointsInPolygon1 into polygon1 starting from idxOfLastPointBeforeDeletionP1
	polygon1.Coords = append(polygon1.Coords[:idxOfLastPointBeforeDeletionP1], pointsInPolygon1...)

	return polygon1
}

// Get points from P1 that are inside P2
func getPointsInPolygon(P1 []dtoschema.CoordinateStruct, P2 []dtoschema.CoordinateStruct) ([]dtoschema.CoordinateStruct, *dtoschema.CoordinateStruct) {
	var pointsInPolygon []dtoschema.CoordinateStruct
	pointsIdxInPolygon := []int{}

	for idx, point := range P1 {
		if isPointInPolygon(point, P2) {
			pointsInPolygon = append(pointsInPolygon, point)
			pointsIdxInPolygon = append(pointsIdxInPolygon, idx)
		}
	}

	if len(pointsIdxInPolygon) == 0 {
		return nil, nil
	}

	if len(pointsIdxInPolygon) == len(P1) {
		return pointsInPolygon, nil
	}

	lastPointBeforeDeletion := &P1[(pointsIdxInPolygon[0]-1+len(P1))%len(P1)]

	return pointsInPolygon, lastPointBeforeDeletion
}

func isPointInPolygon(point dtoschema.CoordinateStruct, polygon []dtoschema.CoordinateStruct) bool {
	// ray-casting algorithm
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
