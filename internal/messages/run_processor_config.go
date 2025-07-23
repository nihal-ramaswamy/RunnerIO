package messages

type DistanceFuncType = func(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64

type RunProcessorConfig struct {
	DistanceFunc            DistanceFuncType
	NumSecondsForCycle      int64
	MetersThresholdForCycle float64
}

func NewRunProcessorConfig(distanceFunc DistanceFuncType, numSecondsForCycle int64, metersThresholdForCycle float64) RunProcessorConfig {
	return RunProcessorConfig{
		DistanceFunc:            distanceFunc,
		NumSecondsForCycle:      numSecondsForCycle,
		MetersThresholdForCycle: metersThresholdForCycle,
	}
}
