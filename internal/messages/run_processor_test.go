package messages

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
)

func distanceForTest(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	dist := math.Sqrt(math.Pow(lat2-lat1, 2) + math.Pow(lng2-lng1, 2))
	return dist
}

func getRunProcessorConfig() RunProcessorConfig {
	return NewRunProcessorConfig(distanceForTest, 0, 3)
}

func sortCoordsFunc(a, b []dtoschema.CoordinateStruct) int {
	if a[0].X < b[0].X {
		return -1
	}
	if a[0].X > b[0].X {
		return 1
	} 
	return 0
}

func checkIfCoordsEqual(result, expected []dtoschema.CoordinateStruct) bool {
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

func checkIfEqual(result, expected []dtoschema.PolygonData) bool {
	if len(result) != len(expected) {
		return false
	}

	resultToRunnerMap := make(map[string][][]dtoschema.CoordinateStruct)
	for _, resultPolygon := range result {
		resultToRunnerMap[resultPolygon.Runner] = append(resultToRunnerMap[resultPolygon.Runner], resultPolygon.Coords)
	}
	expectedToRunnerMap := make(map[string][][]dtoschema.CoordinateStruct)
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
			return false
		}
		for i, resultCoord := range resultCoords {
			expectedCoord := expectedCoords[i]
			if !checkIfCoordsEqual(resultCoord, expectedCoord) {
				return false
			}
		}
	}

	return true
}

func addToLiveLinesData(liveLinesData *[]dtoschema.LiveLinesData, X, Y float64, sender string) {
	currentTime := time.Now()
	newLiveLines := dtoschema.LiveLinesData{
		X:         X,
		Y:         Y,
		Time:      currentTime,
		Sender:    sender,
		GroupCode: "nihal",
	}
	*liveLinesData = append(*liveLinesData, newLiveLines)
}

func prettyPrint(val string, data []dtoschema.PolygonData) {
	fmt.Println(val)
	for _, polygon := range data {
		fmt.Println("Runner: " + polygon.Runner)
		for _, coord := range polygon.Coords {
			fmt.Println(coord.X, coord.Y)
		}
	}
}

// Test two runners with no pre existing polygons.
// The first runner has a live path. The second runner has a live path.
// The second runner eats into the first runner's live path
func TestEatPolygon1(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}
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

	result, err := processData(liveLinesData, []dtoschema.PolygonData{}, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{Runner: "nihal", Coords: []dtoschema.CoordinateStruct{
			{X: 4.0, Y: 0.0},
			{X: 4.0, Y: 2.0},
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 6.0},
			{X: 2.0, Y: 6.0},
			{X: 0.0, Y: 4.0},
			{X: 0.0, Y: 2.0},
			{X: 2.0, Y: 0.0},
		}},
		{Runner: "nihal1", Coords: []dtoschema.CoordinateStruct{
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

	if !checkIfEqual(result, expected) {
		prettyPrint("Expected", expected)
		prettyPrint("Got", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Only a single runner with a live path and a polygon. The polygon and the live path intersect.
func TestEatPolygon2(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")

	polygonData := []dtoschema.PolygonData{}
	p1 := dtoschema.PolygonData{
		Runner: "nihal",
		Coords: []dtoschema.CoordinateStruct{
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

	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{Runner: "nihal", Coords: []dtoschema.CoordinateStruct{
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

	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Only a single runner with a live path and a polygon.
// The polygon and the live path do not intersect.
func TestEatPolygon3(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 6.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 0.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 2.0, 0.0, "nihal")

	polygonData := []dtoschema.PolygonData{}
	p1 := dtoschema.PolygonData{
		Runner: "nihal",
		Coords: []dtoschema.CoordinateStruct{
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
		},
	}
	polygonData = append(polygonData, p1)

	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{Runner: "nihal", Coords: []dtoschema.CoordinateStruct{
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
		}},
		{Runner: "nihal", Coords: []dtoschema.CoordinateStruct{
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

	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// The first runner has one polygon and a live path. The two do not intersect.
// The second runner has no polygon and a live path
// The second runner eats into the first runner's live path
func TestEatPolygon4(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}
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

	polygonData := []dtoschema.PolygonData{}
	p1 := dtoschema.PolygonData{
		Runner: "nihal",
		Coords: []dtoschema.CoordinateStruct{
			{X: -8.0, Y: 0.0},
			{X: -10.0, Y: -2.0},
			{X: -10.0, Y: -4.0},
			{X: -8.0, Y: -6.0},
		},
	}
	polygonData = append(polygonData, p1)

	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{Runner: "nihal", Coords: []dtoschema.CoordinateStruct{
			{X: 4.0, Y: 0.0},
			{X: 4.0, Y: 2.0},
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 6.0},
			{X: 2.0, Y: 6.0},
			{X: 0.0, Y: 4.0},
			{X: 0.0, Y: 2.0},
			{X: 2.0, Y: 0.0},
		}},
		{Runner: "nihal1", Coords: []dtoschema.CoordinateStruct{
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 2.0},
			{X: 6.0, Y: 0.0},
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
			{X: 6.0, Y: 6.0},
		}},
		{Runner: "nihal", Coords: []dtoschema.CoordinateStruct{
			{X: -8.0, Y: 0.0},
			{X: -10.0, Y: -2.0},
			{X: -10.0, Y: -4.0},
			{X: -8.0, Y: -6.0},
		}},
	}

	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Both runners have a polygon and a live path.
// The polygons intersect
// The live paths intersect
func TestEatPolygon5(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}
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

	polygonData := []dtoschema.PolygonData{}
	p1 := dtoschema.PolygonData{
		Runner: "nihal",
		Coords: []dtoschema.CoordinateStruct{
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

	p2 := dtoschema.PolygonData{
		Runner: "nihal1",
		Coords: []dtoschema.CoordinateStruct{
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

	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{Runner: "nihal", Coords: []dtoschema.CoordinateStruct{
			{X: 4.0, Y: 0.0},
			{X: 4.0, Y: 2.0},
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 6.0},
			{X: 2.0, Y: 6.0},
			{X: 0.0, Y: 4.0},
			{X: 0.0, Y: 2.0},
			{X: 2.0, Y: 0.0},
		}},
		{Runner: "nihal1", Coords: []dtoschema.CoordinateStruct{
			{X: 4.0, Y: 4.0},
			{X: 4.0, Y: 2.0},
			{X: 6.0, Y: 0.0},
			{X: 8.0, Y: 0.0},
			{X: 10.0, Y: 2.0},
			{X: 10.0, Y: 4.0},
			{X: 8.0, Y: 6.0},
			{X: 6.0, Y: 6.0},
		}},
		{Runner: "nihal1", Coords: []dtoschema.CoordinateStruct{
			{X: 14.0, Y: 4.0},
			{X: 16.0, Y: 4.0},
			{X: 18.0, Y: 4.0},
			{X: 18.0, Y: 6.0},
			{X: 18.0, Y: 8.0},
			{X: 16.0, Y: 8.0},
			{X: 14.0, Y: 8.0},
			{X: 14.0, Y: 6.0},
		}},
		{Runner: "nihal", Coords: []dtoschema.CoordinateStruct{
			{X: 12.0, Y: 2.0},
			{X: 14.0, Y: 2.0},
			{X: 16.0, Y: 2.0},
			{X: 14.0, Y: 4.0},
			{X: 12.0, Y: 6.0},
			{X: 10.0, Y: 6.0},
		}},
	}

	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// One runner, no polygon, one live path that does not form a polygon
func TestEatPolygon6(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}
	addToLiveLinesData(&liveLinesData, 4.0, 0.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 2.0, "nihal")
	addToLiveLinesData(&liveLinesData, 6.0, 4.0, "nihal")
	addToLiveLinesData(&liveLinesData, 4.0, 6.0, "nihal")

	polygonData := []dtoschema.PolygonData{}
	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{}
	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Three runners
// Live path 1 intersects with live path 2
// Live path 2 intersects with live path 3
func TestEatPolygon7(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}

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

	polygonData := []dtoschema.PolygonData{}
	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{
			Runner: "nihal2",
			Coords: []dtoschema.CoordinateStruct{
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
			Coords: []dtoschema.CoordinateStruct{
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
			Coords: []dtoschema.CoordinateStruct{
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
	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Three runner
// Two overlapping live paths, one does not
// No polygons
func TestEatPolygon8(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}

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

	polygonData := []dtoschema.PolygonData{}
	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{
			Runner: "nihal2",
			Coords: []dtoschema.CoordinateStruct{
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
			Coords: []dtoschema.CoordinateStruct{
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
			Coords: []dtoschema.CoordinateStruct{
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
	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Same as TestEatPolygon8, but with a different order of runners
func TestEatPolygon9(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}

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

	polygonData := []dtoschema.PolygonData{}
	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{
			Runner: "nihal1",
			Coords: []dtoschema.CoordinateStruct{
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
			Coords: []dtoschema.CoordinateStruct{
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
	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Three runners
// Two runners overlap the same runner
func TestEatPolygon10(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}

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

	polygonData := []dtoschema.PolygonData{}
	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{
			Runner: "nihal2",
			Coords: []dtoschema.CoordinateStruct{
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
			Coords: []dtoschema.CoordinateStruct{
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
			Coords: []dtoschema.CoordinateStruct{
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
	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Two runners
// Only overlap one a line
func TestEatPolygon11(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}

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

	polygonData := []dtoschema.PolygonData{}
	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{
			Runner: "nihal1",
			Coords: []dtoschema.CoordinateStruct{
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
			Coords: []dtoschema.CoordinateStruct{
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
	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Two runners
// Only overlap one a point
func TestEatPolygon12(t *testing.T) {
	liveLinesData := []dtoschema.LiveLinesData{}

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

	polygonData := []dtoschema.PolygonData{}
	result, err := processData(liveLinesData, polygonData, getRunProcessorConfig())

	if err != nil {
		t.Errorf("Failed to process data: %s", err)
	}

	expected := []dtoschema.PolygonData{
		{
			Runner: "nihal1",
			Coords: []dtoschema.CoordinateStruct{
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
			Coords: []dtoschema.CoordinateStruct{
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
	if !checkIfEqual(result, expected) {
		prettyPrint("Expected: ", expected)
		prettyPrint("Got: ", result)
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
