package services

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	mongo_schema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/mongodb_schema"
)

func distanceForTest(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	dist := math.Sqrt(math.Pow(lat2-lat1, 2) + math.Pow(lng2-lng1, 2))
	return dist
}

func getRunProcessorConfig() RunProcessorConfig {
	return NewRunProcessorConfig(distanceForTest, false, 0, 3)
}

func sortCoordsFunc(a, b []mongo_schema.CoordinateStruct) int {
	if a[0].X < b[0].X {
		return -1
	}
	if a[0].X > b[0].X {
		return 1
	}
	return 0
}

func sortCoordsFunc2(a, b mongo_schema.CoordinateStruct) int {
	if a.X < b.X {
		return -1
	}
	if a.X > b.X {
		return 1
	}
	return 0
}

func checkIfCoordsEqual(result, expected []mongo_schema.CoordinateStruct) bool {
	if len(result) != len(expected) {
		return false
	}
	for i, resultCoord := range result {
		expectedCoord := expected[i]
		if resultCoord.X != expectedCoord.X || resultCoord.Y != expectedCoord.Y {
			return false
		}
	}
	return true
}

func checkIfEqualPolygonData(result, expected []mongo_schema.RunnerPolygonSchema) bool {
	if len(result) != len(expected) {
		fmt.Printf("Expected %v, got %v\n", len(expected), len(result))
		return false
	}

	resultToRunnerMap := make(map[string][][]mongo_schema.CoordinateStruct)
	for _, resultPolygon := range result {
		resultToRunnerMap[resultPolygon.Runner] = append(resultToRunnerMap[resultPolygon.Runner], resultPolygon.Coords)
	}
	expectedToRunnerMap := make(map[string][][]mongo_schema.CoordinateStruct)
	for _, expectedPolygon := range expected {
		expectedToRunnerMap[expectedPolygon.Runner] = append(expectedToRunnerMap[expectedPolygon.Runner], expectedPolygon.Coords)
	}

	for _, resultCoords := range resultToRunnerMap {
		slices.SortFunc(resultCoords, sortCoordsFunc)
	}

	for _, expectedCoords := range expectedToRunnerMap {
		slices.SortFunc(expectedCoords, sortCoordsFunc)
	}

	for runner, resultCoords := range resultToRunnerMap {
		expectedCoords := expectedToRunnerMap[runner]
		if len(resultCoords) != len(expectedCoords) {
			fmt.Printf("Expected for runner: %v %v, got %v\n", runner, len(expectedCoords), len(resultCoords))
			return false
		}
		for i, resultCoord := range resultCoords {
			expectedCoord := expectedCoords[i]
			if !checkIfCoordsEqual(resultCoord, expectedCoord) {
				fmt.Printf("%v coord not matching for runner %v\n", i, runner)
				return false
			}
		}
	}

	return true
}

func checkIfEqualLiveLinesData(result, expected []mongo_schema.RunnerLiveLinesData) bool {
	if len(result) != len(expected) {
		return false
	}

	resultToRunnerMap := make(map[string][]mongo_schema.CoordinateStruct)
	expectedToRunnerMap := make(map[string][]mongo_schema.CoordinateStruct)

	for _, resultPolygon := range result {
		resultToRunnerMap[resultPolygon.Sender] = append(resultToRunnerMap[resultPolygon.Sender], mongo_schema.CoordinateStruct{
			X: resultPolygon.X,
			Y: resultPolygon.Y,
		})
	}
	for _, expectedPolygon := range expected {
		expectedToRunnerMap[expectedPolygon.Sender] = append(expectedToRunnerMap[expectedPolygon.Sender], mongo_schema.CoordinateStruct{
			X: expectedPolygon.X,
			Y: expectedPolygon.Y,
		})
	}

	for _, resultCoords := range resultToRunnerMap {
		slices.SortFunc(resultCoords, sortCoordsFunc2)
	}
	for _, expectedCoords := range expectedToRunnerMap {
		slices.SortFunc(expectedCoords, sortCoordsFunc2)
	}

	for sender, resultCoords := range resultToRunnerMap {
		expectedCoords := expectedToRunnerMap[sender]
		if len(resultCoords) != len(expectedCoords) {
			return false
		}
		for i, resultCoord := range resultCoords {
			expectedCoord := expectedCoords[i]
			if resultCoord.X != expectedCoord.X || resultCoord.Y != expectedCoord.Y {
				return false
			}
		}
	}

	return true
}

func addToLiveLinesData(liveLinesData *[]mongo_schema.RunnerLiveLinesData, X, Y float64, sender string) {
	currentTime := time.Now()
	newLiveLines := mongo_schema.RunnerLiveLinesData{
		X:         X,
		Y:         Y,
		Time:      currentTime,
		Sender:    sender,
		GroupCode: "nihal",
	}
	*liveLinesData = append(*liveLinesData, newLiveLines)
}

func prettyPrint(val string, data []mongo_schema.RunnerPolygonSchema) {
	fmt.Println(val)
	for _, polygon := range data {
		fmt.Println("Runner: " + polygon.Runner)
		for _, coord := range polygon.Coords {
			fmt.Println(coord.X, coord.Y)
		}
	}
}

func prettyPrint2(val string, data []mongo_schema.RunnerLiveLinesData) {
	fmt.Println(val)
	for _, polygon := range data {
		fmt.Printf("Runner: %v, X: %v, Y: %v\n", polygon.Sender, polygon.X, polygon.Y)
	}
}

// Test two runners with no pre existing polygons.
// The first runner has a live path. The second runner has a live path.
// The second runner eats into the first runner's live path
func TestEatPolygon1(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")

	addToLiveLinesData(&liveLinesData, 4.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 10.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 10.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal1")

	_, result, err := processData(liveLinesData, []mongo_schema.RunnerPolygonSchema{}, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{Runner: "nihal", Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 0.0},
			{X: 4.0, Y: 2.0},
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 6.0},
			{X: 2.0, Y: 6.0},
			{X: 0.0, Y: 4.0},
			{X: 0.0, Y: 2.0},
			{X: 2.0, Y: 0.0},
		}},
		{Runner: "nihal1", Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 2.0},
			{X: 6.0, Y: 0.0},
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
			{X: 6.0, Y: 6.0},
		}},
	}

	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected", expected)
		prettyPrint("Got", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Only a single runner with a live path and a polygon. The polygon and the live path intersect.
func TestEatPolygon2(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	p1 := mongo_schema.RunnerPolygonSchema{
		Runner: "nihal",
		Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 2.0},
			{X: 6.0, Y: 0.0},
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
			{X: 6.0, Y: 6.0},
		},
	}
	polygonData = append(polygonData, p1)

	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{Runner: "nihal", Coords: []mongo_schema.CoordinateStruct{
			{X: 6.0, Y: 0.0},
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
			{X: 6.0, Y: 6.0},
			{X: 4.0, Y: 6.0},
			{X: 2.0, Y: 6.0},
			{X: 0.0, Y: 4.0},
			{X: 0.0, Y: 2.0},
			{X: 2.0, Y: 0.0},
			{X: 4.0, Y: 0.0},
		}},
	}

	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Only a single runner with a live path and a polygon.
// The polygon and the live path do not intersect.
func TestEatPolygon3(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	p1 := mongo_schema.RunnerPolygonSchema{
		Runner: "nihal",
		Coords: []mongo_schema.CoordinateStruct{
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
		},
	}
	polygonData = append(polygonData, p1)

	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{Runner: "nihal", Coords: []mongo_schema.CoordinateStruct{
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
		}},
		{Runner: "nihal", Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 0.0},
			{X: 6.0, Y: 2.0},
			{X: 6.0, Y: 4.0},
			{X: 4.0, Y: 6.0},
			{X: 2.0, Y: 6.0},
			{X: 0.0, Y: 4.0},
			{X: 0.0, Y: 2.0},
			{X: 2.0, Y: 0.0},
		}},
	}

	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// The first runner has one polygon and a live path. The two do not intersect.
// The second runner has no polygon and a live path
// The second runner eats into the first runner's live path
func TestEatPolygon4(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")

	addToLiveLinesData(&liveLinesData, 4.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 10.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 10.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal1")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	p1 := mongo_schema.RunnerPolygonSchema{
		Runner: "nihal",
		Coords: []mongo_schema.CoordinateStruct{
			{X: -8.0, Y: 0.0},
			{X: -10.0, Y: -2.0},
			{X: -10.0, Y: -4.0},
			{X: -8.0, Y: -6.0},
		},
	}
	polygonData = append(polygonData, p1)

	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{Runner: "nihal", Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 0.0},
			{X: 4.0, Y: 2.0},
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 6.0},
			{X: 2.0, Y: 6.0},
			{X: 0.0, Y: 4.0},
			{X: 0.0, Y: 2.0},
			{X: 2.0, Y: 0.0},
		}},
		{Runner: "nihal1", Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 2.0},
			{X: 6.0, Y: 0.0},
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
			{X: 6.0, Y: 6.0},
		}},
		{Runner: "nihal", Coords: []mongo_schema.CoordinateStruct{
			{X: -8.0, Y: 0.0},
			{X: -10.0, Y: -2.0},
			{X: -10.0, Y: -4.0},
			{X: -8.0, Y: -6.0},
		}},
	}

	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Both runners have a polygon and a live path.
// The polygons intersect
// The live paths intersect
func TestEatPolygon5(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")

	addToLiveLinesData(&liveLinesData, 4.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 10.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 10.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal1")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	p1 := mongo_schema.RunnerPolygonSchema{
		Runner: "nihal",
		Coords: []mongo_schema.CoordinateStruct{
			{X: 12.0, Y: 2.0},
			{X: 14.0, Y: 2.0},
			{X: 16.0, Y: 2.0},
			{X: 16.0, Y: 4.0},
			{X: 16.0, Y: 6.0},
			{X: 14.0, Y: 6.0},
			{X: 12.0, Y: 6.0},
			{X: 10.0, Y: 6.0},
		},
		Time: time.Now().Add(time.Second * -1),
	}

	p2 := mongo_schema.RunnerPolygonSchema{
		Runner: "nihal1",
		Coords: []mongo_schema.CoordinateStruct{
			{X: 14.0, Y: 4.0},
			{X: 16.0, Y: 4.0},
			{X: 18.0, Y: 4.0},
			{X: 18.0, Y: 6.0},
			{X: 18.0, Y: 8.0},
			{X: 16.0, Y: 8.0},
			{X: 14.0, Y: 8.0},
			{X: 14.0, Y: 6.0},
		},
		Time: time.Now(),
	}
	polygonData = append(polygonData, p1, p2)

	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{Runner: "nihal", Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 0.0},
			{X: 4.0, Y: 2.0},
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 6.0},
			{X: 2.0, Y: 6.0},
			{X: 0.0, Y: 4.0},
			{X: 0.0, Y: 2.0},
			{X: 2.0, Y: 0.0},
		}},
		{Runner: "nihal1", Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 2.0},
			{X: 6.0, Y: 0.0},
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
			{X: 6.0, Y: 6.0},
		}},
		{Runner: "nihal1", Coords: []mongo_schema.CoordinateStruct{
			{X: 14.0, Y: 4.0},
			{X: 16.0, Y: 4.0},
			{X: 18.0, Y: 4.0},
			{X: 18.0, Y: 6.0},
			{X: 18.0, Y: 8.0},
			{X: 16.0, Y: 8.0},
			{X: 14.0, Y: 8.0},
			{X: 14.0, Y: 6.0},
		}},
		{Runner: "nihal", Coords: []mongo_schema.CoordinateStruct{
			{X: 12.0, Y: 2.0},
			{X: 14.0, Y: 2.0},
			{X: 16.0, Y: 2.0},
			{X: 14.0, Y: 4.0},
			{X: 12.0, Y: 6.0},
			{X: 10.0, Y: 6.0},
		}},
	}

	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// One runner, no polygon, one live path that does not form a polygon
func TestEatPolygon6(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{}
	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Three runners
// Live path 1 intersects with live path 2
// Live path 2 intersects with live path 3
func TestEatPolygon7(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}

	addToLiveLinesData(&liveLinesData, 1.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 3.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 1.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 3.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 1.0, "nihal")

	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 1.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 5.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 1.0, "nihal1")

	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 10.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 11.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 12.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 1.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 2.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 3.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 4.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 5.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 12.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 11.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 10.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 5.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 4.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 3.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 2.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 1.0, "nihal2")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal2",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 8.0, Y: 0.0},
				{X: 9.0, Y: 0.0},
				{X: 10.0, Y: 0.0},
				{X: 11.0, Y: 0.0},
				{X: 12.0, Y: 0.0},
				{X: 13.0, Y: 0.0},
				{X: 13.0, Y: 1.0},
				{X: 13.0, Y: 2.0},
				{X: 13.0, Y: 3.0},
				{X: 13.0, Y: 4.0},
				{X: 13.0, Y: 5.0},
				{X: 13.0, Y: 6.0},
				{X: 12.0, Y: 6.0},
				{X: 11.0, Y: 6.0},
				{X: 10.0, Y: 6.0},
				{X: 9.0, Y: 6.0},
				{X: 8.0, Y: 6.0},
				{X: 8.0, Y: 5.0},
				{X: 8.0, Y: 4.0},
				{X: 8.0, Y: 3.0},
				{X: 8.0, Y: 2.0},
				{X: 8.0, Y: 1.0},
			},
		},
		{
			Runner: "nihal1",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 4.0, Y: 0.0},
				{X: 5.0, Y: 0.0},
				{X: 6.0, Y: 0.0},
				{X: 7.0, Y: 0.0},
				{X: 8.0, Y: 1.0},
				{X: 8.0, Y: 2.0},
				{X: 8.0, Y: 3.0},
				{X: 8.0, Y: 4.0},
				{X: 8.0, Y: 5.0},
				{X: 7.0, Y: 6.0},
				{X: 6.0, Y: 6.0},
				{X: 5.0, Y: 6.0},
				{X: 4.0, Y: 6.0},
				{X: 4.0, Y: 5.0},
				{X: 4.0, Y: 4.0},
				{X: 4.0, Y: 3.0},
				{X: 4.0, Y: 2.0},
				{X: 4.0, Y: 1.0},
			},
		},
		{
			Runner: "nihal",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 1.0, Y: 0.0},
				{X: 2.0, Y: 0.0},
				{X: 3.0, Y: 0.0},
				{X: 4.0, Y: 1.0},
				{X: 4.0, Y: 2.0},
				{X: 4.0, Y: 3.0},
				{X: 4.0, Y: 4.0},
				{X: 4.0, Y: 5.0},
				{X: 3.0, Y: 6.0},
				{X: 2.0, Y: 6.0},
				{X: 1.0, Y: 6.0},
				{X: 1.0, Y: 5.0},
				{X: 1.0, Y: 4.0},
				{X: 1.0, Y: 3.0},
				{X: 1.0, Y: 2.0},
				{X: 1.0, Y: 1.0},
			},
		},
	}
	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Three runner
// Two overlapping live paths, one does not
// No polygons
func TestEatPolygon8(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}

	addToLiveLinesData(&liveLinesData, -1.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, -2.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, -3.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, -4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, -5.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, -6.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, -6.0, 1.0, "nihal")
	addToLiveLinesData(&liveLinesData, -6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, -6.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, -6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, -6.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, -6.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, -5.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, -4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, -3.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, -2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, -1.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, -1.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, -1.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, -1.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, -1.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, -1.0, 1.0, "nihal")

	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 1.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 5.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 1.0, "nihal1")

	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 10.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 11.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 12.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 1.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 2.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 3.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 4.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 5.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 12.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 11.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 10.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 5.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 4.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 3.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 2.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 1.0, "nihal2")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal2",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 8.0, Y: 0.0},
				{X: 9.0, Y: 0.0},
				{X: 10.0, Y: 0.0},
				{X: 11.0, Y: 0.0},
				{X: 12.0, Y: 0.0},
				{X: 13.0, Y: 0.0},
				{X: 13.0, Y: 1.0},
				{X: 13.0, Y: 2.0},
				{X: 13.0, Y: 3.0},
				{X: 13.0, Y: 4.0},
				{X: 13.0, Y: 5.0},
				{X: 13.0, Y: 6.0},
				{X: 12.0, Y: 6.0},
				{X: 11.0, Y: 6.0},
				{X: 10.0, Y: 6.0},
				{X: 9.0, Y: 6.0},
				{X: 8.0, Y: 6.0},
				{X: 8.0, Y: 5.0},
				{X: 8.0, Y: 4.0},
				{X: 8.0, Y: 3.0},
				{X: 8.0, Y: 2.0},
				{X: 8.0, Y: 1.0},
			},
		},
		{
			Runner: "nihal1",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 4.0, Y: 0.0},
				{X: 5.0, Y: 0.0},
				{X: 6.0, Y: 0.0},
				{X: 7.0, Y: 0.0},
				{X: 8.0, Y: 1.0},
				{X: 8.0, Y: 2.0},
				{X: 8.0, Y: 3.0},
				{X: 8.0, Y: 4.0},
				{X: 8.0, Y: 5.0},
				{X: 7.0, Y: 6.0},
				{X: 6.0, Y: 6.0},
				{X: 5.0, Y: 6.0},
				{X: 4.0, Y: 6.0},
				{X: 4.0, Y: 5.0},
				{X: 4.0, Y: 4.0},
				{X: 4.0, Y: 3.0},
				{X: 4.0, Y: 2.0},
				{X: 4.0, Y: 1.0},
			},
		},
		{
			Runner: "nihal",
			Coords: []mongo_schema.CoordinateStruct{
				{X: -1.0, Y: 0.0},
				{X: -2.0, Y: 0.0},
				{X: -3.0, Y: 0.0},
				{X: -4.0, Y: 0.0},
				{X: -5.0, Y: 0.0},
				{X: -6.0, Y: 0.0},
				{X: -6.0, Y: 1.0},
				{X: -6.0, Y: 2.0},
				{X: -6.0, Y: 3.0},
				{X: -6.0, Y: 4.0},
				{X: -6.0, Y: 5.0},
				{X: -6.0, Y: 6.0},
				{X: -5.0, Y: 6.0},
				{X: -4.0, Y: 6.0},
				{X: -3.0, Y: 6.0},
				{X: -2.0, Y: 6.0},
				{X: -1.0, Y: 6.0},
				{X: -1.0, Y: 5.0},
				{X: -1.0, Y: 4.0},
				{X: -1.0, Y: 3.0},
				{X: -1.0, Y: 2.0},
				{X: -1.0, Y: 1.0},
			},
		},
	}
	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Same as TestEatPolygon8, but with a different order of runners
func TestEatPolygon9(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}

	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 1.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 5.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 1.0, "nihal1")

	addToLiveLinesData(&liveLinesData, 1.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 3.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 1.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 3.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 1.0, "nihal")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal1",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 7.0, Y: 0.0},
				{X: 8.0, Y: 0.0},
				{X: 9.0, Y: 0.0},
				{X: 9.0, Y: 1.0},
				{X: 9.0, Y: 2.0},
				{X: 9.0, Y: 3.0},
				{X: 9.0, Y: 4.0},
				{X: 9.0, Y: 5.0},
				{X: 9.0, Y: 6.0},
				{X: 8.0, Y: 6.0},
				{X: 7.0, Y: 6.0},
				{X: 6.0, Y: 5.0},
				{X: 6.0, Y: 4.0},
				{X: 6.0, Y: 3.0},
				{X: 6.0, Y: 2.0},
				{X: 6.0, Y: 1.0},
			},
		},
		{
			Runner: "nihal",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 1.0, Y: 0.0},
				{X: 2.0, Y: 0.0},
				{X: 3.0, Y: 0.0},
				{X: 4.0, Y: 0.0},
				{X: 5.0, Y: 0.0},
				{X: 6.0, Y: 0.0},
				{X: 6.0, Y: 1.0},
				{X: 6.0, Y: 2.0},
				{X: 6.0, Y: 3.0},
				{X: 6.0, Y: 4.0},
				{X: 6.0, Y: 5.0},
				{X: 6.0, Y: 6.0},
				{X: 5.0, Y: 6.0},
				{X: 4.0, Y: 6.0},
				{X: 3.0, Y: 6.0},
				{X: 2.0, Y: 6.0},
				{X: 1.0, Y: 6.0},
				{X: 1.0, Y: 5.0},
				{X: 1.0, Y: 4.0},
				{X: 1.0, Y: 3.0},
				{X: 1.0, Y: 2.0},
				{X: 1.0, Y: 1.0},
			},
		},
	}
	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Three runners
// Two runners overlap the same runner
func TestEatPolygon10(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}

	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 1.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 5.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 4.0, 1.0, "nihal1")

	addToLiveLinesData(&liveLinesData, 1.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 3.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 1.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 3.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 1.0, "nihal")

	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 10.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 11.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 12.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 0.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 1.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 2.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 3.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 4.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 5.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 13.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 12.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 11.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 10.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 5.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 4.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 3.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 2.0, "nihal2")
	addToLiveLinesData(&liveLinesData, 8.0, 1.0, "nihal2")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal2",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 8.0, Y: 0.0},
				{X: 9.0, Y: 0.0},
				{X: 10.0, Y: 0.0},
				{X: 11.0, Y: 0.0},
				{X: 12.0, Y: 0.0},
				{X: 13.0, Y: 0.0},
				{X: 13.0, Y: 1.0},
				{X: 13.0, Y: 2.0},
				{X: 13.0, Y: 3.0},
				{X: 13.0, Y: 4.0},
				{X: 13.0, Y: 5.0},
				{X: 13.0, Y: 6.0},
				{X: 12.0, Y: 6.0},
				{X: 11.0, Y: 6.0},
				{X: 10.0, Y: 6.0},
				{X: 9.0, Y: 6.0},
				{X: 8.0, Y: 6.0},
				{X: 8.0, Y: 5.0},
				{X: 8.0, Y: 4.0},
				{X: 8.0, Y: 3.0},
				{X: 8.0, Y: 2.0},
				{X: 8.0, Y: 1.0},
			},
		},
		{
			Runner: "nihal1",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 7.0, Y: 0.0},
				{X: 8.0, Y: 1.0},
				{X: 8.0, Y: 2.0},
				{X: 8.0, Y: 3.0},
				{X: 8.0, Y: 4.0},
				{X: 8.0, Y: 5.0},
				{X: 7.0, Y: 6.0},
				{X: 6.0, Y: 5.0},
				{X: 6.0, Y: 4.0},
				{X: 6.0, Y: 3.0},
				{X: 6.0, Y: 2.0},
				{X: 6.0, Y: 1.0},
			},
		},
		{
			Runner: "nihal",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 1.0, Y: 0.0},
				{X: 2.0, Y: 0.0},
				{X: 3.0, Y: 0.0},
				{X: 4.0, Y: 0.0},
				{X: 5.0, Y: 0.0},
				{X: 6.0, Y: 0.0},
				{X: 6.0, Y: 1.0},
				{X: 6.0, Y: 2.0},
				{X: 6.0, Y: 3.0},
				{X: 6.0, Y: 4.0},
				{X: 6.0, Y: 5.0},
				{X: 6.0, Y: 6.0},
				{X: 5.0, Y: 6.0},
				{X: 4.0, Y: 6.0},
				{X: 3.0, Y: 6.0},
				{X: 2.0, Y: 6.0},
				{X: 1.0, Y: 6.0},
				{X: 1.0, Y: 5.0},
				{X: 1.0, Y: 4.0},
				{X: 1.0, Y: 3.0},
				{X: 1.0, Y: 2.0},
				{X: 1.0, Y: 1.0},
			},
		},
	}
	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Two runners
// Only overlap one a line
func TestEatPolygon11(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}

	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 1.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 1.0, "nihal1")

	addToLiveLinesData(&liveLinesData, 1.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 3.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 1.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 3.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 5.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 3.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 1.0, 1.0, "nihal")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal1",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 7.0, Y: 0.0},
				{X: 8.0, Y: 0.0},
				{X: 9.0, Y: 0.0},
				{X: 9.0, Y: 1.0},
				{X: 9.0, Y: 2.0},
				{X: 9.0, Y: 3.0},
				{X: 9.0, Y: 4.0},
				{X: 9.0, Y: 5.0},
				{X: 9.0, Y: 6.0},
				{X: 8.0, Y: 6.0},
				{X: 7.0, Y: 6.0},
			},
		},
		{
			Runner: "nihal",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 1.0, Y: 0.0},
				{X: 2.0, Y: 0.0},
				{X: 3.0, Y: 0.0},
				{X: 4.0, Y: 0.0},
				{X: 5.0, Y: 0.0},
				{X: 6.0, Y: 0.0},
				{X: 6.0, Y: 1.0},
				{X: 6.0, Y: 2.0},
				{X: 6.0, Y: 3.0},
				{X: 6.0, Y: 4.0},
				{X: 6.0, Y: 5.0},
				{X: 6.0, Y: 6.0},
				{X: 5.0, Y: 6.0},
				{X: 4.0, Y: 6.0},
				{X: 3.0, Y: 6.0},
				{X: 2.0, Y: 6.0},
				{X: 1.0, Y: 6.0},
				{X: 1.0, Y: 5.0},
				{X: 1.0, Y: 4.0},
				{X: 1.0, Y: 3.0},
				{X: 1.0, Y: 2.0},
				{X: 1.0, Y: 1.0},
			},
		},
	}
	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Two runners
// Only overlap one a point
func TestEatPolygon12(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}

	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 10.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 11.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 12.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 12.0, 7.0, "nihal")
	addToLiveLinesData(&liveLinesData, 12.0, 8.0, "nihal")
	addToLiveLinesData(&liveLinesData, 12.0, 9.0, "nihal")
	addToLiveLinesData(&liveLinesData, 12.0, 10.0, "nihal")
	addToLiveLinesData(&liveLinesData, 12.0, 11.0, "nihal")
	addToLiveLinesData(&liveLinesData, 12.0, 12.0, "nihal")
	addToLiveLinesData(&liveLinesData, 11.0, 12.0, "nihal")
	addToLiveLinesData(&liveLinesData, 10.0, 12.0, "nihal")
	addToLiveLinesData(&liveLinesData, 9.0, 12.0, "nihal")
	addToLiveLinesData(&liveLinesData, 9.0, 11.0, "nihal")
	addToLiveLinesData(&liveLinesData, 9.0, 10.0, "nihal")
	addToLiveLinesData(&liveLinesData, 9.0, 9.0, "nihal")
	addToLiveLinesData(&liveLinesData, 9.0, 8.0, "nihal")
	addToLiveLinesData(&liveLinesData, 9.0, 7.0, "nihal")

	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 1.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 9.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 8.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 7.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 6.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 5.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 3.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal1")
	addToLiveLinesData(&liveLinesData, 6.0, 1.0, "nihal1")

	polygonData := []mongo_schema.RunnerPolygonSchema{}
	_, result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal1",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 6.0, Y: 0.0},
				{X: 7.0, Y: 0.0},
				{X: 8.0, Y: 0.0},
				{X: 9.0, Y: 0.0},
				{X: 9.0, Y: 1.0},
				{X: 9.0, Y: 2.0},
				{X: 9.0, Y: 3.0},
				{X: 9.0, Y: 4.0},
				{X: 9.0, Y: 5.0},
				{X: 9.0, Y: 6.0},
				{X: 8.0, Y: 6.0},
				{X: 7.0, Y: 6.0},
				{X: 6.0, Y: 6.0},
				{X: 6.0, Y: 5.0},
				{X: 6.0, Y: 4.0},
				{X: 6.0, Y: 3.0},
				{X: 6.0, Y: 2.0},
				{X: 6.0, Y: 1.0},
			},
		},
		{
			Runner: "nihal",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 10.0, Y: 6.0},
				{X: 11.0, Y: 6.0},
				{X: 12.0, Y: 6.0},
				{X: 12.0, Y: 7.0},
				{X: 12.0, Y: 8.0},
				{X: 12.0, Y: 9.0},
				{X: 12.0, Y: 10.0},
				{X: 12.0, Y: 11.0},
				{X: 12.0, Y: 12.0},
				{X: 11.0, Y: 12.0},
				{X: 10.0, Y: 12.0},
				{X: 9.0, Y: 12.0},
				{X: 9.0, Y: 11.0},
				{X: 9.0, Y: 10.0},
				{X: 9.0, Y: 9.0},
				{X: 9.0, Y: 8.0},
				{X: 9.0, Y: 7.0},
			},
		},
	}
	if !checkIfEqualPolygonData(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Two runners
// One has polygon, the other has a live path partially overlapping the polygon
func TestEatPolygon13(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 7.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 10.0, 0.0, "nihal")

	polygonData := []mongo_schema.RunnerPolygonSchema{}

	p1 := mongo_schema.RunnerPolygonSchema{
		Runner: "nihal1",
		Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 6.0},
			{X: 5.0, Y: 6.0},
			{X: 7.0, Y: 6.0},
			{X: 7.0, Y: 5.0},
			{X: 7.0, Y: 4.0},
			{X: 7.0, Y: 3.0},
			{X: 7.0, Y: 2.0},
			{X: 7.0, Y: 1.0},
			{X: 7.0, Y: 0.0},
			{X: 7.0, Y: -1.0},
			{X: 7.0, Y: -2.0},
			{X: 7.0, Y: -3.0},
			{X: 6.0, Y: -3.0},
			{X: 5.0, Y: -3.0},
			{X: 4.0, Y: -3.0},
			{X: 4.0, Y: -2.0},
			{X: 4.0, Y: -1.0},
			{X: 4.0, Y: 0.0},
			{X: 4.0, Y: 1.0},
			{X: 4.0, Y: 2.0},
			{X: 4.0, Y: 3.0},
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 5.0},
		},
	}
	polygonData = append(polygonData, p1)
	actualLiveLineData, actualPolygonData, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}
	expectedPolygonData := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal1",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 4.0, Y: 6.0},
				{X: 5.0, Y: 6.0},
				{X: 7.0, Y: 6.0},
				{X: 7.0, Y: 5.0},
				{X: 7.0, Y: 4.0},
				{X: 7.0, Y: 3.0},
				{X: 7.0, Y: 2.0},
				{X: 7.0, Y: 1.0},
				{X: 7.0, Y: 0.0},
				{X: 7.0, Y: -1.0},
				{X: 7.0, Y: -2.0},
				{X: 7.0, Y: -3.0},
				{X: 6.0, Y: -3.0},
				{X: 5.0, Y: -3.0},
				{X: 4.0, Y: -3.0},
				{X: 4.0, Y: -2.0},
				{X: 4.0, Y: -1.0},
				{X: 4.0, Y: 0.0},
				{X: 4.0, Y: 1.0},
				{X: 4.0, Y: 2.0},
				{X: 4.0, Y: 3.0},
				{X: 4.0, Y: 4.0},
				{X: 4.0, Y: 5.0},
			},
		},
	}
	expectedLiveLineData := []mongo_schema.RunnerLiveLinesData{
		{X: 8.0, Y: 0.0, Sender: "nihal"},
		{X: 9.0, Y: 0.0, Sender: "nihal"},
		{X: 10.0, Y: 0.0, Sender: "nihal"},
	}

	if !checkIfEqualPolygonData(actualPolygonData, expectedPolygonData) {
		prettyPrint("Expected: ", expectedPolygonData)
		prettyPrint("Got: ", actualPolygonData)
		t.Errorf("Expected %v, got %v", expectedPolygonData, actualPolygonData)
	}

	if !checkIfEqualLiveLinesData(actualLiveLineData, expectedLiveLineData) {
		prettyPrint2("Expected: ", expectedLiveLineData)
		prettyPrint2("Got: ", actualLiveLineData)
		t.Errorf("Expected %v, got %v", expectedLiveLineData, actualLiveLineData)
	}
}

// Two runners
// One has polygon, the other has a live path partially overlapping the polygon
func TestEatPolygon14(t *testing.T) {
	liveLinesData := []mongo_schema.RunnerLiveLinesData{}
	addToLiveLinesData(&liveLinesData, 1.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 3.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 5.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 7.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 8.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 9.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 10.0, 0.0, "nihal")

	polygonData := []mongo_schema.RunnerPolygonSchema{}

	p1 := mongo_schema.RunnerPolygonSchema{
		Runner: "nihal1",
		Coords: []mongo_schema.CoordinateStruct{
			{X: 4.0, Y: 6.0},
			{X: 5.0, Y: 6.0},
			{X: 7.0, Y: 6.0},
			{X: 7.0, Y: 5.0},
			{X: 7.0, Y: 4.0},
			{X: 7.0, Y: 3.0},
			{X: 7.0, Y: 2.0},
			{X: 7.0, Y: 1.0},
			{X: 7.0, Y: 0.0},
			{X: 7.0, Y: -1.0},
			{X: 7.0, Y: -2.0},
			{X: 7.0, Y: -3.0},
			{X: 6.0, Y: -3.0},
			{X: 5.0, Y: -3.0},
			{X: 4.0, Y: -3.0},
			{X: 4.0, Y: -2.0},
			{X: 4.0, Y: -1.0},
			{X: 4.0, Y: 0.0},
			{X: 4.0, Y: 1.0},
			{X: 4.0, Y: 2.0},
			{X: 4.0, Y: 3.0},
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 5.0},
		},
	}
	polygonData = append(polygonData, p1)
	actualLiveLineData, actualPolygonData, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}
	expectedPolygonData := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal1",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 4.0, Y: 6.0},
				{X: 5.0, Y: 6.0},
				{X: 7.0, Y: 6.0},
				{X: 7.0, Y: 5.0},
				{X: 7.0, Y: 4.0},
				{X: 7.0, Y: 3.0},
				{X: 7.0, Y: 2.0},
				{X: 7.0, Y: 1.0},
				{X: 7.0, Y: 0.0},
				{X: 7.0, Y: -1.0},
				{X: 7.0, Y: -2.0},
				{X: 7.0, Y: -3.0},
				{X: 6.0, Y: -3.0},
				{X: 5.0, Y: -3.0},
				{X: 4.0, Y: -3.0},
				{X: 4.0, Y: -2.0},
				{X: 4.0, Y: -1.0},
				{X: 4.0, Y: 0.0},
				{X: 4.0, Y: 1.0},
				{X: 4.0, Y: 2.0},
				{X: 4.0, Y: 3.0},
				{X: 4.0, Y: 4.0},
				{X: 4.0, Y: 5.0},
			},
		},
	}
	expectedLiveLineData := []mongo_schema.RunnerLiveLinesData{
		{X: 8.0, Y: 0.0, Sender: "nihal"},
		{X: 9.0, Y: 0.0, Sender: "nihal"},
		{X: 10.0, Y: 0.0, Sender: "nihal"},
	}

	if !checkIfEqualPolygonData(actualPolygonData, expectedPolygonData) {
		prettyPrint("Expected: ", expectedPolygonData)
		prettyPrint("Got: ", actualPolygonData)
		t.Errorf("Expected %v, got %v", expectedPolygonData, actualPolygonData)
	}

	if !checkIfEqualLiveLinesData(actualLiveLineData, expectedLiveLineData) {
		prettyPrint2("Expected: ", expectedLiveLineData)
		prettyPrint2("Got: ", actualLiveLineData)
		t.Errorf("Expected %v, got %v", expectedLiveLineData, actualLiveLineData)
	}
}

// Two groups running in parallel
// Groups do not intersect

// Two groups running in parallel
// Groups intersect
