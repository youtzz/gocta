package main

import "fmt"

func main() {
	//err := internal.Evaluate(internal.EngineCfg{
	//	Strategy:       internal.StrategyTemplate{},
	//	DataRepo:       internal.NewData(),
	//	Symbol:         "BTCUSDT",
	//	Start:          time.Date(2022, 1, 1, 0, 0, 0, 0, time.Local),
	//	BackTestingMod: internal.BAR,
	//})
	//if err != nil {
	//	panic(err)
	//}

	m := map[string]int{}
	m["1"] = 1
	m["2"] = 2

	var list []int
	for _, v := range m {
		if v == 1 {
			v = 10
		} else if v == 2 {
			v = 12
		}
		list = append(list, v)
	}
	fmt.Println(m)
	fmt.Println(list)
}
