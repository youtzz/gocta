package pkg

import (
	"gocta/internal"
	"testing"
	"time"
)

func TestBollStrategy(t *testing.T) {
	boll := new(BollStrategy)

	// BTCUSDT 1mbar 2021-01-01 until now
	internal.Evaluate(internal.EngineCfg{
		Strategy:       boll,
		DataRepo:       internal.NewData(),
		Symbol:         "BTCUSDT",
		Start:          time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local),
		Interval:       time.Minute,
		BackTestingMod: internal.BAR,
	})
}
