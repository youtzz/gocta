package internal

import (
	expb "newgitlab.com/xquant/exchange-protocols/protocols/src/go"
	"time"
)

type BarData expb.BarData
type TickData expb.TickData
type OrderData expb.OrderData
type TradeData expb.TradeData

type Exchange expb.Exchange

type Interval time.Duration

func PbToInterval(i expb.Interval) Interval {
	switch i {
	case expb.Interval_MINUTES:
		return Interval(time.Minute)
	default:
		panic("not support interval")
	}
}

type Strategy interface {
	//GetCfg() StrategyConfig

	OnInit()
	OnStart()
	OnTick(tick TickData)
	OnBar(bar BarData)
	OnStop()
}

type StrategyConfig struct {
	Name   string
	Author string
}
