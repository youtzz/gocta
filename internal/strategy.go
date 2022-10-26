package internal

import (
	"fmt"
)

type StrategyTemplate struct{}

func (s StrategyTemplate) OnInit() {
	fmt.Println("Strategy: OnInit")
}

func (s StrategyTemplate) OnStart() {
	fmt.Println("Strategy: OnStart")
}

func (s StrategyTemplate) OnTick(tick TickData) {
	fmt.Println("Strategy: OnTick")
}

func (s StrategyTemplate) OnBar(bar BarData) {
	fmt.Printf("Strategy: OnBar: %+v\n", bar)
}

func (s StrategyTemplate) OnStop() {
	fmt.Println("Strategy: OnStop")
}

func (s StrategyTemplate) OnPosition(posChange float64) {
	fmt.Printf("Strategy: OnPosition: changed: %d\n", posChange)
}

func (s StrategyTemplate) OnOrder(order OrderData) {
	fmt.Printf("Strategy: OnOrder: %+v\n", order)
}

func (s StrategyTemplate) OnTrade(trade TradeData) {
	fmt.Printf("Strategy: OnTrade: %+v\n", trade)
}
