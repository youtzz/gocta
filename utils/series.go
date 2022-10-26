package utils

import "github.com/go-gota/gota/series"

func SubSeries(a, b series.Series, name string) series.Series {
	if a.Type() != b.Type() {
		return series.Series{}
	}
	aSlice := a.Float()
	bSlice := b.Float()
	result := make([]float64, len(aSlice))
	for i := range aSlice {
		result[i] = aSlice[i] - bSlice[i]
	}
	return series.New(result, series.Float, name)
}

func DivSeries(a, b series.Series, name string) series.Series {
	if a.Type() != b.Type() {
		return series.Series{}
	}
	aSlice := a.Float()
	bSlice := b.Float()
	result := make([]float64, len(aSlice))
	for i := range aSlice {
		result[i] = aSlice[i] / bSlice[i]
	}
	return series.New(result, series.Float, name)
}

func MutiSeries(a series.Series, muti float64, name string) series.Series {
	if name == "" {
		name = a.Name
	}
	aSlice := a.Float()
	result := make([]float64, len(aSlice))
	for i := range aSlice {
		result[i] = aSlice[i] * muti
	}
	return series.New(result, series.Float, name)
}

func CumSumSeries(a series.Series, name string, capital float64) series.Series {
	if name == "" {
		name = a.Name
	}
	aSlice := a.Float()
	result := make([]float64, len(aSlice))
	var temp float64
	for i, v := range aSlice {
		temp += v
		result[i] = temp + capital
	}
	return series.New(result, series.Float, name)
}

func GreaterSeries(a series.Series, greater float64) (count int) {
	aSlice := a.Float()
	for _, v := range aSlice {
		if v > greater {
			count ++
		}
	}
	return
}

func LowerSeries(a series.Series, lower float64) (count int) {
	aSlice := a.Float()
	for _, v := range aSlice {
		if v < lower {
			count ++
		}
	}
	return
}

func ShiftSeries(a series.Series, shift int, name string) series.Series {
	if name == "" {
		name = a.Name
	}
	aSlice := a.Float()
	var temp int = 0 + shift
	result := make([]float64, len(aSlice))
	for _, v := range aSlice {
		if temp < 0 {
			temp ++
			continue
		}
		if temp > len(aSlice) - 1 {
			break
		}
		result[temp] = v
		temp++
	}
	return series.New(result, series.Float, name)
}