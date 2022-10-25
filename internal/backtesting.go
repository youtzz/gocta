package internal

import (
	"container/list"
	"errors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	expb "newgitlab.com/xquant/exchange-protocols/protocols/src/go"
	"strconv"
	"time"
)

func Evaluate(cfg EngineCfg) error {
	engine, err := newEngine(cfg)
	if err != nil {
		engine.logger.Println("创建引擎失败:", err.Error())
		return err
	}

	err = engine.loadData()
	if err != nil {
		engine.logger.Println("加载数据失败:", err.Error())
		return err
	}

	engine.runBackTesting()

	return nil
}

type BackTestingMod int

const (
	BAR = BackTestingMod(iota)
	TICK
)

type DataRepo interface {
	GetBarData(symbol string, exchange Exchange, interval time.Duration, start, end time.Time) (*list.List, error)
	GetTickData(symbol string, exchange Exchange, interval time.Duration, start, end time.Time) (*list.List, error)
}

type EngineCfg struct {
	Strategy
	DataRepo

	Symbol     string
	Start, End time.Time
	Interval   time.Duration
	Exchange   Exchange

	BackTestingMod
}

func (c EngineCfg) Check() error {
	if c.Symbol == "" {
		return errors.New("symbol不能为空")
	}
	if c.Strategy == nil {
		return errors.New("strategy不能为空")
	}
	if c.DataRepo == nil {
		return errors.New("dataRepo不能为空")
	}

	return nil
}

type Logger interface {
	Println(v ...any)
}

type defaultLogger struct {
}

func (defaultLogger) Println(v ...any) {
	log.Println(v...)
}

type BackTestingEngine struct {
	EngineCfg

	logger Logger

	historyData *list.List

	bar      BarData
	tick     TickData
	datetime time.Time

	activeLimitOrders map[string]*OrderData
	trades            map[string]TradeData
	tradeCount        int
}

func newEngine(cfg EngineCfg) (*BackTestingEngine, error) {
	if err := cfg.Check(); err != nil {
		return nil, err
	}
	return &BackTestingEngine{
		EngineCfg:         cfg,
		logger:            defaultLogger{},
		historyData:       list.New(),
		activeLimitOrders: make(map[string]*OrderData),
		trades:            make(map[string]TradeData),
	}, nil
}

func (b *BackTestingEngine) runBackTesting() {
	b.Strategy.OnInit()
	// todo 缺了让策略提前获得部分交易数据的功能
	b.logger.Println("策略初始化完成")

	b.Strategy.OnStart()
	b.logger.Println("开始回放历史数据")

	// todo 没有回放进度的功能
	for cur := b.historyData.Front(); cur.Next() != nil; cur = cur.Next() {
		if b.BackTestingMod == BAR {
			b.newBar(cur.Value.(BarData))
		} else {
			b.newTick(cur.Value.(TickData))
		}
	}

	b.Strategy.OnStop()
	b.logger.Println("历史数据回放结束")
}

func (b *BackTestingEngine) clear() {

}

func (b *BackTestingEngine) loadData() error {
	b.logger.Println("开始加载历史数据")

	if b.End.IsZero() {
		b.End = time.Now()
	}

	if !b.Start.Before(b.End) {
		return errors.New("起始日期必须小于结束日期")
	}

	b.historyData.Init()

	totalDays := int(b.End.Sub(b.Start).Hours() / 24)
	progressDays := Max(totalDays/10, 1)
	progressDelta := time.Duration(progressDays * 24)
	intervalDelta := b.Interval

	start := b.Start
	end := b.End.Add(progressDelta)
	progress := 0

	var howToLoad func(string, Exchange, time.Duration, time.Time, time.Time) (*list.List, error)
	if b.BackTestingMod == BAR {
		howToLoad = b.DataRepo.GetBarData
	} else {
		howToLoad = b.DataRepo.GetTickData
	}

	for start.Before(end) {
		loaded, err := howToLoad(b.Symbol, b.Exchange, b.Interval, start, end)
		if err != nil {
			return err
		}
		b.historyData.PushBackList(loaded)

		progress += progressDays / totalDays
		progress = Min(progress, 1)

		start = end.Add(intervalDelta)
		end = end.Add(progressDelta)
	}

	b.logger.Println("历史数据加载完成，数据量:", b.historyData.Len())

	return nil
}

func (b *BackTestingEngine) newBar(bar BarData) {
	b.bar = bar
	b.datetime = bar.UpdatedAt.AsTime()

	b.crossLimitOrder()
	//b.crossStopOrder()
	b.Strategy.OnBar(bar)
}

func (b *BackTestingEngine) newTick(tick TickData) {

}

func (b *BackTestingEngine) crossLimitOrder() {
	var (
		longCrossPrice  float64
		shortCrossPrice float64
		longBestPrice   float64
		shortBestPrice  float64
	)
	if b.BackTestingMod == BAR {
		longCrossPrice = b.bar.LowPrice
		shortCrossPrice = b.bar.HighPrice
		longBestPrice = b.bar.OpenPrice
		shortBestPrice = b.bar.OpenPrice
	} else {
		longCrossPrice = b.tick.AskPrice_1
		shortCrossPrice = b.tick.BidPrice_1
		longBestPrice = longCrossPrice
		shortBestPrice = shortCrossPrice
	}

	for _, order := range b.activeLimitOrders {
		longCross := (order.Direction == expb.Direction_LONG) &&
			(order.Price >= longCrossPrice) &&
			(longCrossPrice > 0)

		shortCross := (order.Direction == expb.Direction_SHORT) &&
			(order.Price <= shortCrossPrice) &&
			(shortCrossPrice > 0)

		if !longCross || !shortCross {
			continue
		}

		order.Traded = order.Volume
		order.Status = expb.Status_ALL_TRADED

		b.Strategy.OnOrder(*order)
		if _, ok := b.activeLimitOrders[order.OrderNo]; ok {
			delete(b.activeLimitOrders, order.OrderNo)
		}

		b.tradeCount++

		var (
			tradePrice float64
			posChange  float64
		)

		if longCross {
			tradePrice = Min(order.Price, longBestPrice)
			posChange = order.Volume
		} else {
			tradePrice = Min(order.Price, shortBestPrice)
			posChange = -order.Volume
		}

		trade := TradeData{
			Symbol:    order.Symbol,
			Exchange:  order.Exchange,
			OrderNo:   order.OrderNo,
			TradeNo:   strconv.Itoa(b.tradeCount),
			Direction: order.Direction,
			Offset:    order.Offset,
			Price:     tradePrice,
			Volume:    order.Volume,
			UpdatedAt: timestamppb.New(b.datetime),
			Reference: "",
			//GatewayName: 0,
		}

		b.Strategy.OnPosition(posChange)
		b.Strategy.OnTrade(trade)

		b.trades[trade.TradeNo] = trade
	}
}
