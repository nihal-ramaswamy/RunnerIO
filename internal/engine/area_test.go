package services

import (
	"testing"

	mongo_schema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/mongodb_schema"
)

func TestArea1(t *testing.T) {
	polygonData := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 0, Y: 0},
				{X: 0, Y: 1},
				{X: 1, Y: 1},
				{X: 1, Y: 0},
			},
		},
	}
	expected := 1.0

	actual := getPolygonArea(polygonData[0].Coords)
	if actual != expected {
		t.Errorf("Expected %f, got %f", expected, actual)
	}
}

func TestArea2(t *testing.T) {
	polygonData := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal",
			Coords: []mongo_schema.CoordinateStruct{
				{X: 0, Y: 0},
				{X: 1, Y: -1},
				{X: 2, Y: 0},
				{X: 3, Y: 1},
				{X: 2, Y: 2},
				{X: 1, Y: 3},
				{X: 0, Y: 2},
				{X: -1, Y: 1},
			},
		},
	}
	expected := 8.0

	actual := getPolygonArea(polygonData[0].Coords)
	if actual != expected {
		t.Errorf("Expected %f, got %f", expected, actual)
	}
}

func TestArea3(t *testing.T) {
	polygonData := []mongo_schema.RunnerPolygonSchema{
		{
			Runner: "nihal",
			Coords: []mongo_schema.CoordinateStruct{
				{X: -2, Y: -2},
				{X: 0, Y: 0},
				{X: 0, Y: 2},
				{X: 2, Y: 0},
				{X: 0, Y: 4},
			},
		},
	}
	expected := 6.0

	actual := getPolygonArea(polygonData[0].Coords)
	if actual != expected {
		t.Errorf("Expected %f, got %f", expected, actual)
	}
}
