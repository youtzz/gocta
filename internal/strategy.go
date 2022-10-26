package internal

import (
	"log"
)

type StrategyTemplate struct{}

func (s StrategyTemplate) OnInit() {
	log.Println("Strategy: OnInit")
}

func (s StrategyTemplate) OnStart() {
	log.Println("Strategy: OnStart")
}

func (s StrategyTemplate) OnTick(tick TickData) {
	log.Println("Strategy: OnTick")
}

func (s StrategyTemplate) OnBar(bar BarData) {
	log.Printf("Strategy: OnBar: %+v\n", bar)
}

func (s StrategyTemplate) OnStop() {
	log.Println("Strategy: OnStop")
}

func (s StrategyTemplate) OnPosition(posChange float64) {
	log.Printf("Strategy: OnPosition: changed: %d\n", posChange)
}

func (s StrategyTemplate) OnOrder(order OrderData) {
	log.Printf("Strategy: OnOrder: %+v\n", order)
}

func (s StrategyTemplate) OnTrade(trade TradeData) {
	log.Printf("Strategy: OnTrade: %+v\n", trade)
}
