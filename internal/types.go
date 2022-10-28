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

	OnPosition(posChange float64)
	OnOrder(order OrderData)
	OnTrade(trade TradeData)
}

type StrategyConfig struct {
	Name   string
	Author string
}

type DailyResult struct {
	Date       string
	ClosePrice float64
	PreClose   float64
	Trades     []*TradeData
	TradeCount int
	StartPos   float64
	EndPos     float64
	Turnover   float64
	Commission float64
	Slippage   float64
	TradingPnl float64
	HoldingPnl float64
	TotalPnl   float64
	NetPnl     float64
}

func NewDailyResult(date string, price float64) *DailyResult {
	return &DailyResult{Date: date, ClosePrice: price, Trades: make([]*TradeData, 0)}
}

func (d *DailyResult) AddTrade(trade *TradeData) {
	d.Trades = append(d.Trades, trade)
}

func (d *DailyResult) CalculatePnl(preClose, startPos, rate, Slippage, size float64, inverse bool) {
	if preClose != 0 {
		d.PreClose = preClose
	} else {
		d.PreClose = 1
	}

	d.StartPos = startPos
	d.EndPos = startPos

	if inverse {
		d.HoldingPnl = d.StartPos * (1/d.PreClose - 1/d.ClosePrice) * size
	} else {
		d.HoldingPnl = d.StartPos * (d.ClosePrice - d.PreClose) * size
	}

	d.TradeCount = len(d.Trades)

	for _, trade := range d.Trades {
		var posChange float64
		if trade.Direction == expb.Direction_LONG {
			posChange = trade.Volume
		} else {
			posChange = -trade.Volume
		}
		d.EndPos += posChange

		var turnover float64
		if inverse {
			turnover = trade.Volume * size / trade.Price
			d.TradingPnl += posChange * (1/trade.Price - 1/d.ClosePrice) * size
			d.Slippage += trade.Volume * size * Slippage / (trade.Price * 2) // todo: **2
		} else {
			turnover = trade.Volume * size * trade.Price
			d.TradingPnl += posChange * (d.ClosePrice - trade.Price) * size
			d.Slippage += trade.Volume * size * Slippage
		}
		d.Turnover += turnover
		d.Commission += turnover * rate
	}
	d.TotalPnl = d.TradingPnl + d.HoldingPnl
	d.NetPnl = d.TotalPnl - d.Commission - d.Slippage
}
