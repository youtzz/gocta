package main

import (
	"fmt"
	"gocta/internal"
)

type Strategy struct {
}

func (s Strategy) OnInit() {
	fmt.Println("OnInit")
}

func (s Strategy) OnStart() {
	fmt.Println("OnStart")
}

func (s Strategy) OnTick(tick internal.TickData) {
	fmt.Println("OnTick")
}

func (s Strategy) OnBar(bar internal.BarData) {
	fmt.Println("get Bar ", bar)
}

func (s Strategy) OnStop() {
	fmt.Println("OnStop")
}

func (s Strategy) OnPosition(posChange float64) {
	fmt.Println("OnPosition")
}

func (s Strategy) OnOrder(order internal.OrderData) {
	fmt.Println("OnOrder")
}

func (s Strategy) OnTrade(trade internal.TradeData) {
	fmt.Println("OnTrade")
}
