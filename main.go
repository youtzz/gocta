package main

import (
	"gocta/internal"
	"time"
)

func main() {
	err := internal.Evaluate(internal.EngineCfg{
		Strategy:       Strategy{},
		DataRepo:       internal.NewData(),
		Symbol:         "BTCUSDT",
		Start:          time.Date(2022, 1, 1, 0, 0, 0, 0, time.Local),
		BackTestingMod: internal.BAR,
	})
	if err != nil {
		panic(err)
	}
}
