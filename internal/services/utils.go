package services

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"

	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func putData[T dtoschema.TestGenerics](
	ctx context.Context,
	mongoClient *mongo.Client,
	groupCode string,
	log *zap.Logger,
	collectionName string,
	data []T) ([]interface{}, error) {
	dataInterface := []interface{}{}
	for _, d := range data {
		dataInterface = append(dataInterface, d)
	}
	results, err := mongoClient.Database(
		constants.RUNNER_DATABASE).Collection(
		collectionName).InsertMany(
		ctx, dataInterface)

	if err != nil {
		log.Error("Failed to insert data", zap.Error(err))
	}
	return []interface{}{results}, err

}

func getData[T dtoschema.TestGenerics](
	ctx context.Context,
	mongoClient *mongo.Client,
	groupCode string,
	log *zap.Logger,
	collectionName string,
	currentMaxTime int) ([]T, error) {
	cursor, err := mongoClient.Database(
		constants.RUNNER_DATABASE).Collection(
		collectionName).Find(
		ctx, bson.M{"group_code": groupCode, "inserted_time": bson.M{"$gte": currentMaxTime}})

	if err != nil {
		log.Error("Failed to find document", zap.Error(err))
		return nil, err
	}

	type dataStruct struct {
		Data T `bson:"data"`
	}

	data := []T{}
	maxTime := -1

	for cursor.Next(ctx) {
		var d bson.M
		if err := cursor.Decode(&d); err != nil {
			log.Error("Failed to decode document", zap.Error(err))
			continue
		}
		log.Info("Processing document", zap.Any("data", d))

		dByte, err := bson.MarshalExtJSON(d, true, true)
		if err != nil {
			log.Error("Failed to marshal document", zap.Error(err))
			continue
		}

		var groupData dataStruct
		if err := bson.UnmarshalExtJSON(dByte, true, &groupData); err != nil {
			log.Error("Failed to unmarshal document", zap.Error(err))
			continue
		}

		data = append(data, groupData.Data)
		maxTime = max(maxTime, groupData.Data.GetInsertedTime())
	}

	dataWithMaxTime := make([]T, 0)
	for _, d := range data {
		if d.GetInsertedTime() < maxTime {
			continue
		}
		dataWithMaxTime = append(dataWithMaxTime, d)
	}

	return dataWithMaxTime, nil
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

/*
isCycle checks if the live lines data forms a cycle.

A cycle is formed when the first and last point of the live lines data are within a distance of half a meter and the time difference is more than 10 seconds.
*/
func IsCycle(liveLines *[]dtoschema.LiveLinesData, runProcessorConfig *RunProcessorConfig) bool {
	if len(*liveLines) < 2 {
		return false
	}

	firstPoint := (*liveLines)[0]
	lastPoint := (*liveLines)[len(*liveLines)-1]

	// Rule for a cycle: The first and last point should have a distance of less than half a meter and the time difference should be more than 10 seconds
	if runProcessorConfig.DoSecondsForCycleCheck && firstPoint.Time.Sub(lastPoint.Time).Seconds() < float64(runProcessorConfig.NumSecondsForCycle)-constants.DELTA {
		return false
	}

	if runProcessorConfig.DistanceFunc(firstPoint.X, firstPoint.Y, lastPoint.X, lastPoint.Y) >
		runProcessorConfig.MetersThresholdForCycle-constants.DELTA {
		return false
	}
	return true
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

func getCurrentMaxTime(ctx context.Context, redisClient *redis.Client, groupCode string, log *zap.Logger) int {
	startTime := time.Unix(0, 0).Nanosecond()
	key := fmt.Sprintf("%s:%s", constants.REDIS_CURRENT_MAX_TIME, groupCode)
	maxTimeStr, err := redisClient.Get(ctx, key).Result()

	if err != nil {
		log.Error("Failed to get current max time", zap.Error(err))
		return startTime
	}
	if maxTimeStr == "" {
		return startTime
	}
	maxTimeInt, err := strconv.Atoi(maxTimeStr)
	if err != nil {
		log.Error("Failed to convert max time to int", zap.Error(err))
		return startTime
	}

	return maxTimeInt

}
